'''
# PentAGI Guardian: Guía de Integración y Despliegue

**Versión:** 1.0.0
**Fecha:** 2026-02-19

## 1. Introducción

**PentAGI Guardian** es una extensión para PentAGI que añade capacidades de auto-remediación de seguridad. Este documento proporciona una guía completa para integrar y desplegar el módulo Guardian en un entorno de producción de PentAGI.

### 1.1. Arquitectura General

Guardian se integra en la arquitectura existente de PentAGI de la siguiente manera:

-   **Backend (Go):** Un nuevo paquete `guardian` que contiene la lógica de negocio, los agentes de IA y las herramientas de remediación.
-   **Base de Datos (PostgreSQL):** Nuevas tablas para gestionar tareas de Guardian, planes de remediación, pasos, análisis de vulnerabilidades y backups.
-   **API (GraphQL):** Nuevos tipos, queries, mutations y subscriptions para interactuar con el módulo Guardian desde el frontend.
-   **Frontend (React + TypeScript):** Nuevas páginas y componentes para visualizar y gestionar las tareas de Guardian.

## 2. Requisitos Previos

-   Una instancia de PentAGI funcional (versión compatible).
-   Acceso a la base de datos PostgreSQL de PentAGI.
-   Permisos para modificar la configuración y el código de PentAGI.
-   Go (versión 1.21 o superior) y Node.js (versión 20 o superior) instalados en el entorno de desarrollo.

## 3. Pasos de Integración

### 3.1. Aplicar Migraciones de Base de Datos

El primer paso es aplicar las nuevas migraciones de base de datos para crear las tablas necesarias para Guardian.

1.  **Asegúrate de tener una copia de seguridad de tu base de datos.**
2.  Establece la variable de entorno `DATABASE_URL` con la cadena de conexión a tu base de datos PostgreSQL.
3.  Ejecuta el siguiente comando desde el directorio raíz del proyecto:

    ```bash
    make guardian-migrate
    ```

    Esto ejecutará el archivo de migración `backend/migrations/sql/20260219_210000_guardian_tables.sql` y creará las tablas y tipos necesarios.

### 3.2. Configurar Variables de Entorno

Añade las siguientes variables de entorno a tu archivo `.env` o al sistema de gestión de configuración que utilices:

```env
# --- Guardian Configuration ---

# Activa o desactiva el módulo Guardian
GUARDIAN_ENABLED=true

# Requiere aprobación manual para cada plan de remediación
GUARDIAN_REQUIRE_APPROVAL=true

# Ejecuta en modo de simulación (no se aplican cambios reales)
GUARDIAN_DRY_RUN=false

# Número máximo de reintentos para una tarea fallida
GUARDIAN_MAX_RETRIES=3

# Timeout en segundos para las tareas de Guardian
GUARDIAN_TIMEOUT=1800

# Directorio para almacenar los backups
GUARDIAN_BACKUP_DIR=/var/pentagi/backups/guardian

# Días de retención para los backups
GUARDIAN_BACKUP_RETENTION_DAYS=30

# --- Fin de la Configuración de Guardian ---
```

### 3.3. Generar Modelos de Datos SQLC

Si has realizado cambios en los archivos SQL de `backend/sqlc/models/`, necesitas regenerar los modelos de datos de Go con SQLC.

1.  Asegúrate de tener `sqlc` instalado (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`).
2.  Ejecuta el siguiente comando desde el directorio `backend/`:

    ```bash
    sqlc generate
    ```

### 3.4. Recompilar la Aplicación

Ahora, necesitas recompilar la aplicación PentAGI para incluir el nuevo código de Guardian.

1.  **Backend:**

    ```bash
    cd backend
    go build -o bin/pentagi ./cmd/pentagi
    ```

2.  **Frontend:**

    ```bash
    cd frontend
    pnpm install
    pnpm build
    ```

## 4. Despliegue

### 4.1. Actualizar la Imagen de Docker

Si utilizas Docker para el despliegue, necesitas reconstruir tu imagen de Docker para incluir los cambios.

1.  Asegúrate de que tu `Dockerfile` copia los nuevos archivos y ejecuta los comandos de build.
2.  Reconstruye la imagen:

    ```bash
    docker build -t vxcontrol/pentagi:latest-guardian .
    ```

### 4.2. Actualizar la Configuración de Docker Compose

Si utilizas `docker-compose`, asegúrate de que el servicio `pentagi` utiliza la nueva imagen y que las variables de entorno de Guardian están configuradas correctamente.

```yaml
services:
  pentagi:
    image: vxcontrol/pentagi:latest-guardian
    environment:
      - GUARDIAN_ENABLED=true
      - GUARDIAN_REQUIRE_APPROVAL=true
      # ... resto de variables de Guardian
```

### 4.3. Desplegar la Nueva Versión

Despliega la nueva versión de la aplicación como lo harías normalmente (ej. `docker-compose up -d`).

## 5. Verificación Post-Despliegue

1.  **Accede a la interfaz de PentAGI.**
2.  **Verifica que el nuevo menú "Guardian" aparece en la barra lateral.**
3.  **Crea una nueva tarea de Guardian** para escanear un sitio web en busca de cabeceras de seguridad faltantes.
4.  **Verifica que se crea un plan de remediación** y que puedes aprobarlo o rechazarlo.
5.  **Aprueba el plan** y verifica que se ejecuta correctamente y que los cambios se aplican en el servidor de destino (si no estás en modo `DryRun`).
6.  **Revisa los logs** de Guardian en busca de errores.

## 6. Troubleshooting

-   **Error "guardian.view permission denied":** Asegúrate de que los roles de usuario tienen los permisos de Guardian asignados en la tabla `privileges`.
-   **Las tareas de Guardian no se ejecutan:** Verifica que el servicio de PentAGI tiene acceso a Docker y puede crear contenedores.
-   **Los backups no se crean:** Asegúrate de que el directorio `GUARDIAN_BACKUP_DIR` existe y que el usuario de PentAGI tiene permisos de escritura en él.

---
*Fin del documento.*
'''
