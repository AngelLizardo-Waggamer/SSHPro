# SSHPro

CLI interactiva (TUI) para gestionar hosts SSH y conectarte rápidamente usando el comando nativo `ssh`.

## Requisitos
- Go **1.26+**
- Cliente `ssh` disponible en el PATH (OpenSSH en Windows o sistemas Unix)

## Instalación
```bash
go build -o sshpro .
```

Opcionalmente:
```bash
go install
```

## Uso básico
```bash
./sshpro
```

Al iniciar:
- Se crea automáticamente `~/.ssh_configured_hosts.json` si no existe.
- Se carga la lista de hosts y se muestra el menú interactivo.

## Atajos
- `[enter]` conectar
- `[/]` buscar
- `[a]` añadir
- `[e]` editar
- `[d]` eliminar (con confirmación)
- `[q]` salir

## Búsqueda y filtros
Pulsa `/` para entrar en modo búsqueda y escribir el texto a filtrar.
- `esc` cancela la búsqueda
- Mientras se está filtrando, los demás atajos se ignoran para evitar acciones accidentales.

## Configuración (JSON)

### Hosts
Ruta por defecto:
- `~/.ssh_configured_hosts.json`


Formato esperado (array de objetos):
```json
[
  {
    "name": "Server1",
    "user": "root",
    "ip": "111.111.111.111",
    "key": "id_ed25519",
    "port": 22
  }
]
```

Notas:
- `port` es opcional (default: `22`).
- `key` vacío implica usar la llave por defecto del usuario (`~/.ssh/id_rsa` o `~/.ssh/id_ed25519`). Para el caso de **servidores que están protegidos por contraseña** y no tienen una llave configurada para acceder también se debe de dejar vacío.

#### Seguridad y permisos
El archivo se guarda con permisos **0600** (solo lectura/escritura para el usuario).

### Temas
Se pueden añadir temas personalizados en `~/.sshpro/themes.json`. El formato esperado de un tema es el siguiente:

```json
[
  {
    "name": "blue",
    "colors": {
      "container_border": "33",
      "title": "39",
      "help": "110",
      "status": "81",
      "subtitle": "244",
      "focused_input": "81",
      "blurred_input": "250",
      "error_message": "203",
      "form_title": "39",
      "form_border": "33",
      "selected_border": "33",
      "selected_title": "81",
      "selected_desc": "75",
      "filter_prompt": "75",
      "filter_text": "81"
    }
  }
]
```

## Resolución de llaves SSH
El campo `key` es flexible:
- Si es **ruta absoluta**, se usa tal cual.
- Si empieza con `~`, se resuelve al home del usuario.
- Si es **relativa**, se asume dentro de `~/.ssh/`.

Ejemplos válidos:
- `id_ed25519`
- `keys/servidor-prod`
- `~/.ssh/id_rsa`
- `/home/user/.ssh/id_ed25519`

## Licencia
MIT — ver [`LICENCE.md`](LICENCE.md).

## Disclaimer
Este proyecto es una herramienta personal y no se garantiza compatibilidad con todos los entornos o configuraciones.
