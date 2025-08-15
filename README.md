# JDB - Go Database Library

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/Version-v0.0.6-orange.svg)](https://github.com/celsiainternet/jdb/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-celsiainternet%2Fjdb-black.svg)](https://github.com/celsiainternet/jdb)

JDB es una librería de Go que proporciona una interfaz unificada y simplificada para trabajar con múltiples bases de datos. Ofrece soporte para PostgreSQL, MySQL, SQLite y Oracle con una API consistente y fácil de usar.

## 🆕 Últimas Actualizaciones (v0.0.6)

- **Dependencias actualizadas**: Elvis v1.1.111, Chi router v5.2.2
- **Drivers mejorados**: MySQL v1.9.3, PostgreSQL v1.10.9, SQLite v1.37.1
- **Performance**: Optimizaciones en el manejo de conexiones
- **Estabilidad**: Correcciones de bugs y mejoras en el sistema de daemon
- **Compatibilidad**: Soporte completo para Go 1.23.0+

## 🚀 Características

### 🗄️ Multi-Driver Support

- **PostgreSQL**: Driver nativo con soporte completo para características avanzadas
- **MySQL**: Integración con go-sql-driver/mysql para máximo rendimiento
- **SQLite**: Soporte con modernc.org/sqlite para aplicaciones embebidas
- **Oracle**: Driver especializado para entornos empresariales

### 🏗️ Arquitectura Moderna

- **API Unificada**: Interfaz consistente independientemente del motor de base de datos
- **ORM Simplificado**: Definición declarativa de modelos y esquemas
- **CQRS Ready**: Soporte integrado para Command Query Responsibility Segregation
- **Core System**: Sistema de metadatos y gestión automática de modelos

### ⚡ Performance & Scale

- **Transacciones**: Soporte completo para transacciones ACID
- **Bulk Operations**: Operaciones masivas optimizadas
- **Connection Pooling**: Gestión automática de conexiones
- **Query Optimization**: Optimización automática de consultas

### 🔧 Developer Experience

- **Debug Mode**: Sistema de depuración avanzado para desarrollo
- **Type Safety**: Tipado fuerte con validaciones automáticas
- **Hot Reload**: Recarga automática de configuraciones
- **JavaScript VM**: Integración con Goja para scripts dinámicos

### 🛠️ DevOps Features

- **Sistema de Daemon**: Gestión completa de servicios con control de ciclo de vida
- **Gestión de PID**: Control automático de procesos
- **Health Checks**: Verificación de estado en tiempo real
- **Graceful Shutdown**: Cierre controlado con manejo de señales

### 🔐 Security & Management

- **Gestión de Usuarios**: Creación y administración de usuarios de base de datos
- **Auditoría**: Sistema de auditoría automática para compliance
- **Eventos**: Hooks antes y después de operaciones para logging y validación
- **Configuration Management**: Configuración dinámica en tiempo de ejecución

### 🔧 Utilidades Integradas

- **ID Generation**: Soporte para ULID, UUID, XID y Snowflake IDs
- **Cache & Redis**: Integración con Redis para caching distribuido
- **Message Queue**: Soporte para NATS messaging
- **Compression**: Algoritmos de compresión integrados
- **Cryptography**: Funciones criptográficas avanzadas

## 📦 Instalación

```bash
go get github.com/celsiainternet/jdb
```

### Dependencias Principales

```bash
# Dependencia principal
go get github.com/celsiainternet/elvis@v1.1.113

# Drivers de base de datos incluidos
# - PostgreSQL: github.com/lib/pq v1.10.9
# - MySQL: github.com/go-sql-driver/mysql v1.9.3
# - SQLite: modernc.org/sqlite v1.37.1
# - HTTP Router: github.com/go-chi/chi/v5 v5.2.2
# - Utilidades adicionales: ULID, UUID, Redis, NATS
```

## 🔧 Configuración

### Variables de Entorno

```bash
# Configuración básica
NODEID=1
DB_NAME=myapp
DB_DRIVER=postgres  # postgres, mysql, sqlite, oracle
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
APP_NAME=myapp

# Configuración adicional
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# Configuración Oracle específica
ORA_DB_SERVICE_NAME_ORACLE=jdb
ORA_DB_SSL_ORACLE=false
ORA_DB_SSL_VERIFY_ORACLE=false
ORA_DB_VERSION_ORACLE=19
```

## 📖 Uso Básico

### Conexión a Base de Datos

```go
package main

import (
    "fmt"

    "github.com/celsiainternet/elvis/et"
    jdb "github.com/celsiainternet/jdb/jdb"
    _ "github.com/celsiainternet/jdb/drivers/postgres" // Importar driver específico
)

func main() {
    // Configuración de conexión
    params := jdb.ConnectParams{
        Driver:   "postgres",
        Name:     "myapp",
        UserCore: true,
        NodeId:   1,
        Debug:    true,
        Params: et.Json{
            "host":     "localhost",
            "port":     5432,
            "username": "postgres",
            "password": "password",
            "database": "myapp",
        },
    }

    // Conectar a la base de datos
    db, err := jdb.ConnectTo(params)
    if err != nil {
        panic(err)
    }
    defer db.Disconected()

    fmt.Println("Conectado a:", db.Name)
}
```

### Definición de Modelos

```go
// Definir un esquema
schema := db.GetSchema("public")

// Definir un modelo
user := schema.DefineModel("users", "Usuarios del sistema")
user.DefineColumn("id", jdb.TypeDataKey, jdb.PrimaryKey)
user.DefineColumn("name", jdb.TypeDataText, jdb.Required)
user.DefineColumn("email", jdb.TypeDataText, jdb.Unique)
user.DefineColumn("age", jdb.TypeDataInt)
user.DefineColumn("active", jdb.TypeDataBool, jdb.Default(true))
user.DefineColumn("created_at", jdb.TypeDataTime, jdb.Default("NOW()"))

// Campos especiales del sistema
user.DefineCreatedAtField()    // Campo de fecha de creación
user.DefineUpdatedAtField()    // Campo de fecha de actualización
user.DefineStatusField()       // Campo de estado
user.DefineSystemKeyField()    // Campo de clave del sistema
user.DefineIndexField()        // Campo de índice
user.DefineSourceField()       // Campo de origen
user.DefineProjectField()      // Campo de proyecto

// Crear el modelo en la base de datos
err := db.LoadModel(user)
if err != nil {
    panic(err)
}
```

### Operaciones CRUD

```go
import (
    "github.com/celsiainternet/elvis/et"
    jdb "github.com/celsiainternet/jdb/jdb"
)

// Insertar datos
result, err := db.Command(&jdb.Command{
    Command: jdb.Insert,
    From:    user.GetFrom(),
    Values: []et.Json{
        {
            "name":  "Juan Pérez",
            "email": "juan@example.com",
            "age":   30,
        },
    },
})

// Consultar datos
items, err := db.Select(&jdb.Ql{
    From: user.GetFrom(),
    Where: &jdb.QlWhere{
        And: []*jdb.Where{
            {Field: "active", Op: jdb.Eq, Value: true},
        },
    },
})

// Actualizar datos
result, err := db.Command(&jdb.Command{
    Command: jdb.Update,
    From:    user.GetFrom(),
    Values: []et.Json{
        {"age": 31},
    },
    QlWhere: &jdb.QlWhere{
        And: []*jdb.Where{
            {Field: "id", Op: jdb.Eq, Value: "user123"},
        },
    },
})

// Eliminar datos
result, err := db.Command(&jdb.Command{
    Command: jdb.Delete,
    From:    user.GetFrom(),
    QlWhere: &jdb.QlWhere{
        And: []*jdb.Where{
            {Field: "id", Op: jdb.Eq, Value: "user123"},
        },
    },
})
```

### Bulk Insert

```go
// Inserción masiva
result, err := db.Command(&jdb.Command{
    Command: jdb.Bulk,
    From:    user.GetFrom(),
    Data: []et.Json{
        {"name": "Ana García", "email": "ana@example.com", "age": 25},
        {"name": "Carlos López", "email": "carlos@example.com", "age": 35},
        {"name": "María Rodríguez", "email": "maria@example.com", "age": 28},
    },
})
```

### Transacciones

```go
// Iniciar transacción
tx, err := db.Begin()
if err != nil {
    panic(err)
}
defer tx.Rollback()

// Operaciones en transacción
result, err := tx.Command(&jdb.Command{
    Command: jdb.Insert,
    From:    user.GetFrom(),
    Values: []et.Json{
        {"name": "Usuario Transaccional", "email": "tx@example.com"},
    },
})

// Commit de la transacción
err = tx.Commit()
if err != nil {
    panic(err)
}
```

## 🛠️ Sistema de Daemon

JDB incluye un sistema de daemon robusto para gestionar servicios con control completo del ciclo de vida:

### Características del Daemon

- **Gestión de PID**: Control automático de archivos PID para evitar múltiples instancias
- **Servidor HTTP**: Servidor web integrado con Chi router
- **Gestión de señales**: Manejo graceful de SIGINT y SIGTERM
- **Control de estado**: Verificación en tiempo real del estado del servicio
- **Configuración dinámica**: Configuración en tiempo de ejecución

### Gestión del Servicio

```bash
# Mostrar ayuda
./jdb help

# Mostrar versión del daemon
./jdb version

# Verificar estado del servicio
./jdb status

# Configurar el servicio (JSON)
./jdb conf '{"port": 3500, "debug": true}'

# Iniciar el servicio en segundo plano
./jdb start

# Detener el servicio gracefully
./jdb stop

# Reiniciar el servicio
./jdb restart
```

### Estructura del Daemon

El daemon utiliza:

- **Archivo PID**: `./tmp/myservice.pid` para control de procesos
- **Interfaz HTTP**: Servidor web en el puerto configurado
- **Logs estructurados**: Sistema de logging integrado con Elvis
- **Configuración JSON**: Parámetros dinámicos en tiempo de ejecución

### Configuración del Daemon

```go
// Ejemplo de configuración programática del daemon
import (
    "github.com/celsiainternet/elvis/et"
    jdb "github.com/celsiainternet/jdb/cmd/jdb"
)

// Configuración del daemon
config := et.Json{
    "port":  3500,
    "debug": true,
    "host":  "localhost",
}

// El daemon se configura automáticamente basado en variables de entorno
// o mediante el comando: ./jdb conf '{"port": 3500, "debug": true}'
```

## 👥 Gestión de Usuarios

JDB proporciona funcionalidades para gestionar usuarios de base de datos:

### PostgreSQL

```go
// Crear usuario
err := db.CreateUser("nuevo_usuario", "password123", "password123")

// Cambiar contraseña
err := db.ChangePassword("nuevo_usuario", "nueva_password", "nueva_password")

// Otorgar privilegios
err := db.GrantPrivileges("nuevo_usuario", "myapp")

// Eliminar usuario
err := db.DeleteUser("nuevo_usuario")
```

### MySQL

```go
// Crear usuario
err := db.CreateUser("nuevo_usuario", "password123", "password123")

// Cambiar contraseña
err := db.ChangePassword("nuevo_usuario", "nueva_password", "nueva_password")

// Otorgar privilegios
err := db.GrantPrivileges("nuevo_usuario", "myapp")

// Eliminar usuario
err := db.DeleteUser("nuevo_usuario")
```

## 🎯 Nuevas Funcionalidades

### JavaScript VM Integration

```go
// Ejecutar scripts JavaScript en el modelo
user.vm.Set("customFunction", func(data et.Json) et.Json {
    // Lógica personalizada
    return data
})

// Ejecutar script
result, err := user.vm.RunString(`
    var data = {name: "Juan", age: 30};
    customFunction(data);
`)
```

### Sistema de Eventos Avanzado

```go
// Definir eventos personalizados
user.On("custom_event", func(message event.Message) {
    console.Log("Evento personalizado:", message)
})

// Emitir eventos
user.Emit("custom_event", event.Message{
    Type: "user_created",
    Data: et.Json{"user_id": "123"},
})
```

### Generación de Datos de Prueba

```go
// Generar datos de prueba para el modelo
testData := user.New("name", "email", "age")
// Resultado: {"name": "", "email": "", "age": 0}

// Generar datos con valores por defecto
testData := user.New()
// Resultado: {"id": "users:ulid", "name": "", "email": "", "age": 0, "active": true, "created_at": "2024-01-01T00:00:00Z"}
```

### Consultas Avanzadas

```go
// Consulta con campos ocultos
items, err := db.Select(&jdb.Ql{
    From: user.GetFrom(),
    Hidden: []string{"password", "secret_key"},
})

// Consulta con datos de origen
items, err := db.Select(&jdb.Ql{
    From: user.GetFrom(),
    TypeSelect: jdb.Source,
})
```

## 🏗️ Estructura del Proyecto

```
jdb/
├── jdb/                 # Paquete principal
│   ├── database.go      # Gestión de conexiones
│   ├── model.go         # Definición de modelos
│   ├── command.go       # Comandos CRUD
│   ├── ql.go           # Query Language
│   ├── model-new.go     # Generación de datos
│   ├── model-define.go  # Definición de campos especiales
│   └── ...
├── drivers/            # Drivers de base de datos
│   ├── postgres/       # Driver PostgreSQL
│   │   ├── users.go    # Gestión de usuarios
│   │   └── ...
│   ├── mysql/          # Driver MySQL
│   │   ├── users.go    # Gestión de usuarios
│   │   └── ...
│   ├── sqlite/         # Driver SQLite
│   ├── oracle/         # Driver Oracle
│   │   ├── users.go    # Gestión de usuarios
│   │   └── ...
├── cqrs/              # Patrón CQRS
└── cmd/               # Aplicación de ejemplo
    ├── jdb/           # Comando principal
    │   ├── main.go     # Punto de entrada
    │   ├── systemd.go  # Sistema de daemon
    │   ├── pid.go      # Gestión de PID
    │   └── msg.go      # Mensajes del sistema
    └── main.go         # Ejemplo de uso
```

## 🔌 Drivers Soportados

### PostgreSQL

```go
import _ "github.com/celsiainternet/jdb/drivers/postgres"

params := jdb.ConnectParams{
    Driver: "postgres",
    Params: et.Json{
        "host":     "localhost",
        "port":     5432,
        "username": "postgres",
        "password": "password",
        "database": "myapp",
        "app":      "myapp",
    },
}
```

### MySQL

```go
import _ "github.com/celsiainternet/jdb/drivers/mysql"

params := jdb.ConnectParams{
    Driver: "mysql",
    Params: et.Json{
        "host":     "localhost",
        "port":     3306,
        "username": "root",
        "password": "password",
        "database": "myapp",
    },
}
```

### SQLite

```go
import _ "github.com/celsiainternet/jdb/drivers/sqlite"

params := jdb.ConnectParams{
    Driver: "sqlite",
    Params: et.Json{
        "database": "./data.db",
    },
}
```

### Oracle

```go
import _ "github.com/celsiainternet/jdb/drivers/oracle"

params := jdb.ConnectParams{
    Driver: "oracle",
    Params: et.Json{
        "host":         "localhost",
        "port":         1521,
        "username":     "system",
        "password":     "password",
        "app":          "myapp",
        "service_name": "XE",
        "ssl":          false,
        "ssl_verify":   false,
        "version":      19,
    },
}
```

## 🎯 Ejemplos Avanzados

### Consultas Complejas

```go
// Consulta con JOIN
items, err := db.Select(&jdb.Ql{
    From: user.GetFrom(),
    Joins: []*jdb.QlJoin{
        {
            Type:  jdb.InnerJoin,
            Table: "profiles",
            On: &jdb.QlWhere{
                And: []*jdb.Where{
                    {Field: "users.id", Op: jdb.Eq, Value: "profiles.user_id"},
                },
            },
        },
    },
    Where: &jdb.QlWhere{
        And: []*jdb.Where{
            {Field: "users.active", Op: jdb.Eq, Value: true},
            {Field: "profiles.verified", Op: jdb.Eq, Value: true},
        },
    },
    OrderBy: &jdb.QlOrder{
        Asc: []*jdb.Field{{Name: "users.created_at"}},
    },
    Limit: 10,
})
```

### Eventos y Hooks

```go
// Evento antes de insertar
user.EventsInsert = append(user.EventsInsert, func(model *jdb.Model, before, after jdb.Json) error {
    fmt.Println("Insertando usuario:", after)
    return nil
})

// Evento después de actualizar
user.EventsUpdate = append(user.EventsUpdate, func(model *jdb.Model, before, after jdb.Json) error {
    fmt.Println("Usuario actualizado:", after)
    return nil
})
```

### Campos Especiales

```go
// Definir campo de texto completo
user.DefineFullText("spanish", []string{"name", "description"})

// Definir relación
user.DefineRelation("profile", "profiles", map[string]string{
    "user_id": "id",
}, 1)

// Definir rollup
user.DefineRollup("total_orders", "orders", map[string]string{
    "user_id": "id",
}, "amount")

// Definir objeto
user.DefineObject("address", "addresses", map[string]string{
    "user_id": "id",
}, []string{"street", "city", "country"})
```

## 🚀 Compilación y Ejecución

### Ejecutar en modo desarrollo

```bash
# Compilar y ejecutar con race detection
gofmt -w . && go run --race ./cmd/jdb

# Ejecutar con parámetros específicos
go run ./cmd/jdb status
go run ./cmd/jdb start
```

### Compilar para producción

```bash
# Compilación optimizada
gofmt -w . && go build -a -o ./jdb ./cmd/jdb

# Ejecutar el binario compilado
./jdb help
./jdb start
```

### Ejemplo con aplicación personalizada

```bash
# Usar cmd/main.go como ejemplo base
go run ./cmd/main.go
```

### Gestión de Versiones Automática

```bash
# Incrementar versión de revisión (X.Y.Z+1)
./version.sh --v

# Incrementar versión menor (X.Y+1.0)
./version.sh --n

# Incrementar versión mayor (X+1.0.0)
./version.sh --m

# Solo crear tag sin commit
./version.sh --version
```

## 📚 API Reference

### Información de Versión

**Versión Actual**: v0.0.6

El sistema de versionado es automático y sigue el estándar semántico (SemVer):

- **Major**: Cambios incompatibles en la API
- **Minor**: Nuevas funcionalidades compatibles hacia atrás
- **Patch**: Correcciones de bugs compatibles

```bash
# Para desarrolladores: proceso de release
git add .
git commit -m 'Update version'
./version.sh --v  # Incrementa patch
git push origin --tags
```

### Tipos de Datos Soportados

- `TypeDataText` - VARCHAR(250)
- `TypeDataShortText` - VARCHAR(80)
- `TypeDataMemo` - TEXT
- `TypeDataInt` - INTEGER
- `TypeDataNumber` - DECIMAL(18,2)
- `TypeDataBool` - BOOLEAN
- `TypeDataTime` - TIMESTAMP
- `TypeDataObject` - JSONB
- `TypeDataArray` - JSONB
- `TypeDataKey` - VARCHAR(80)
- `TypeDataState` - VARCHAR(20)
- `TypeDataSerie` - BIGINT
- `TypeDataPrecision` - DOUBLE PRECISION
- `TypeDataBytes` - BYTEA
- `TypeDataGeometry` - JSONB
- `TypeDataFullText` - TSVECTOR

### Tipos de ID Soportados

- `TpNodeId` - ID de nodo
- `TpUUId` - UUID
- `TpULId` - ULID
- `TpXId` - XID

### Operadores de Consulta

- `Eq` - Igual
- `Ne` - No igual
- `Gt` - Mayor que
- `Gte` - Mayor o igual que
- `Lt` - Menor que
- `Lte` - Menor o igual que
- `Like` - Como
- `ILike` - Como (case insensitive)
- `In` - En
- `NotIn` - No en
- `IsNull` - Es nulo
- `IsNotNull` - No es nulo

### Comandos del Sistema

- `Insert` - Insertar
- `Update` - Actualizar
- `Delete` - Eliminar
- `Bulk` - Inserción masiva
- `Upsert` - Insertar o actualizar
- `Delsert` - Eliminar e insertar

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.

## 🆘 Soporte

Si tienes alguna pregunta o necesitas ayuda, por favor:

1. Revisa la documentación
2. Busca en los issues existentes
3. Crea un nuevo issue con detalles del problema

---

**JDB** - Simplificando el acceso a bases de datos en Go 🚀
