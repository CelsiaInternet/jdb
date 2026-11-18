#!/bin/bash

set -e                                                        # Detener la ejecución en caso de error

# Las versiones son tags anotados vX.Y.Z. Un tag pertenece al commit donde se crea y llega a otras ramas con el
# merge (develop → main). Por eso:
#   - La nueva versión se calcula desde el tag MÁS ALTO de todo el repo (incluye el remoto), no desde el más
#     cercano a HEAD (git describe), para no repetir ni retroceder versiones cuando las ramas tienen tags distintos.
#   - La versión de una rama es el tag más alto que ya llegó a ella por merge (la misma que usa deploy.sh).

HELP=false
MAYOR=false
MINOR=false
REQUEST=false
SHOW=false
BRANCH=$(git branch --show-current)                           # Rama actual

# Parsear opciones
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --h | --help) HELP=true ;;                            # Muestra la ayuda
        --m | --major) MAYOR=true ;;                          # Incrementa la versión mayor
        --n | --minor) MINOR=true ;;                          # Incrementa la versión menor
        --v | --r | --request) REQUEST=true ;;                # Incrementa la revisión (patch)
        --s | --show) SHOW=true ;;                            # Muestra la versión de cada rama
        *) echo "Opción desconocida: $1"; exit 1 ;;
    esac
    shift
done

# Tag de versión más alto alcanzable desde una referencia (por defecto HEAD)
branch_version() {
  local ref=${1:-HEAD}
  local tag
  tag=$(git tag --merged "$ref" -l 'v[0-9]*' --sort=-v:refname | head -n 1)
  echo "${tag:-v0.0.0}"
}

# Tag de versión más alto de todo el repo, en cualquier rama
latest_version() {
  local tag
  tag=$(git tag -l 'v[0-9]*' --sort=-v:refname | head -n 1)
  echo "${tag:-v0.0.0}"
}

# Trae los tags del remoto para no reutilizar una versión creada desde otra rama u otra máquina
fetch_tags() {
  if ! git fetch origin --tags --quiet; then
    echo "No se pudieron traer los tags de origin; se calcula con los tags locales."
  fi
}

show_versions() {
  echo "Versión más alta del repo: $(latest_version)"
  echo "Versión de $BRANCH (HEAD): $(branch_version HEAD)"
  local b
  for b in main develop; do
    if git rev-parse --verify --quiet "origin/$b" > /dev/null; then
      echo "Versión de origin/$b: $(branch_version "origin/$b")"
    fi
  done
}

build_version() {
  local base
  base=$(latest_version)

  if [ "$base" == "v0.0.0" ]; then
    NEW_VERSION="v1.0.0"
    return
  fi

  # Divide la etiqueta en componentes usando el punto como delimitador
  IFS='.' read -r -a version_parts <<< "${base#v}"

  if [ "$MAYOR" == true ]; then
    version_parts[0]=$((version_parts[0] + 1))
    version_parts[1]=0
    version_parts[2]=0
  elif [ "$MINOR" == true ]; then
    version_parts[1]=$((version_parts[1] + 1))
    version_parts[2]=0
  else
    version_parts[2]=$((version_parts[2] + 1))
  fi

  NEW_VERSION="v${version_parts[0]}.${version_parts[1]}.${version_parts[2]}"
}

# Cambia la versión del comando de instalación (…@vX.Y.Z) por la nueva en README.md y README.es.md, si existen.
# Solo toca el patrón @vX.Y.Z para no cambiar otras menciones de versiones (por ejemplo, "desde v1.1.324").
update_readmes() {
  local file
  for file in README.md README.es.md; do
    [ -f "$file" ] || continue
    if grep -Eq '@v[0-9]+\.[0-9]+\.[0-9]+' "$file"; then
      if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' -E "s/@v[0-9]+\.[0-9]+\.[0-9]+/@$NEW_VERSION/g" "$file"
      else
        sed -i -E "s/@v[0-9]+\.[0-9]+\.[0-9]+/@$NEW_VERSION/g" "$file"
      fi
      echo "$file: versión actualizada a $NEW_VERSION (sin commit)"
    fi
  done
}

update_version() {
  # Si HEAD ya tiene versión (por ejemplo, llegó con un merge fast-forward) no se crea otra
  local head_tag
  head_tag=$(git tag --points-at HEAD -l 'v[0-9]*' --sort=-v:refname | head -n 1)
  if [ -n "$head_tag" ]; then
    echo "HEAD ya tiene la versión $head_tag; no se crea un tag nuevo."
    exit 1
  fi

  CURRENT_VERSION=$(branch_version HEAD)
  echo "Rama: $BRANCH"
  echo "Versión actual de la rama: $CURRENT_VERSION"
  echo "Versión más alta del repo: $(latest_version)"
  echo "Nueva versión: $NEW_VERSION"

  update_readmes

  git tag -a "$NEW_VERSION" -m "$NEW_VERSION"
  git push origin "$NEW_VERSION"                              # Solo el tag nuevo, no todos los locales

  echo "Etiqueta $NEW_VERSION creada y enviada a Git"
}

help() {
  echo "Uso: ./version.sh [opciones]"
  echo "Crea el tag de la siguiente versión sobre HEAD, a partir del tag más alto del repo."
  echo "La versión llega a las demás ramas con el merge (develop → main)."
  echo ""
  echo "Opciones:"
  echo "  --h, --help               Muestra este mensaje de ayuda"
  echo "  --m, --major              Incrementa la versión mayor"
  echo "  --n, --minor              Incrementa la versión menor"
  echo "  --v, --r, --request       Incrementa la revisión (patch)"
  echo "  --s, --show               Muestra la versión de la rama actual, de main y de develop"
  exit 0
}

# Mostrar las opciones elegidas
echo "Opciones elegidas:"
[[ "$MAYOR" == true ]] && echo " - Major: Activado"
[[ "$MINOR" == true ]] && echo " - Minor: Activado"
[[ "$REQUEST" == true ]] && echo " - Revisión: Activado"
[[ "$SHOW" == true ]] && echo " - Mostrar versiones: Activado"

if [ "$HELP" == true ]; then
  help
elif [ "$SHOW" == true ]; then
  fetch_tags
  show_versions
elif [ "$MAYOR" == true ] || [ "$MINOR" == true ] || [ "$REQUEST" == true ]; then
  fetch_tags
  build_version
  update_version
else
  help
fi

# Línea en blanco al final
