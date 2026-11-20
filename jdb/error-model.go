package jdb

import (
	"fmt"
	"strings"
)

/**
* modelLabel: Nombre del modelo para los errores: schema.tabla (o solo la tabla sin schema).
* @param model *Model
* @return string
**/
func modelLabel(model *Model) string {
	if model == nil {
		return ""
	}

	if model.Schema == "" {
		return model.Table
	}

	return fmt.Sprintf("%s.%s", model.Schema, model.Table)
}

/**
* describe: Modelos del from para los errores, con su alias: "projects.users (a), projects.teams (b)".
* @return string
**/
func (s *QlFroms) describe() string {
	if s == nil {
		return ""
	}

	result := []string{}
	for _, from := range s.Froms {
		if from == nil || from.Model == nil {
			continue
		}

		label := modelLabel(from.Model)
		if from.As != "" {
			label = fmt.Sprintf("%s (%s)", label, from.As)
		}
		result = append(result, label)
	}

	return strings.Join(result, ", ")
}

/**
* withModel: Antepone al error el modelo donde ocurrio ("projects.users (a): pq: ..."). Usa %w, asi que
* errors.Is y errors.As siguen viendo el error original (p. ej. *pq.Error). No repite el prefijo si
* el error ya lo trae (un comando que hace una consulta del mismo modelo).
* @param label string, err error
* @return error
**/
func withModel(label string, err error) error {
	if err == nil || label == "" {
		return err
	}

	prefix := label + ": "
	if strings.HasPrefix(err.Error(), prefix) {
		return err
	}

	return fmt.Errorf("%s%w", prefix, err)
}
