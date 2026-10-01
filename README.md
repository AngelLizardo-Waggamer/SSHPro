# SSHPro

CLI interactiva (TUI) y cliente rápido por línea de comandos para gestionar hosts SSH y conectarte rápidamente usando el comando nativo `ssh`.

## Requisitos
- Go **1.26+**
- Cliente `ssh` y `scp` disponible en el PATH (OpenSSH en Windows o sistemas Unix)

## Instalación
```bash
go build -o sshpro .
```

Opcionalmente:
```bash
go install
```

---

## Funcionalidades en Modo TUI (Interfaz Interactiva)

Al ejecutar `./sshpro` sin argumentos, se inicia la interfaz interactiva en terminal:

```bash
./sshpro
```

### Características de la TUI:
- **Gestión visual de servidores:** Explora la lista de tus hosts con información clara de usuario, IP/host y clave asociada.
- **Búsqueda y filtrado en tiempo real:** Pulsa `/` para filtrar al instante tu lista de servidores. Mientras buscas, los atajos de edición quedan protegidos para evitar acciones involuntarias (`esc` para salir del filtro).
- **Conexión interactiva:** Al seleccionar un host y presionar `enter`, se lanza una sesión SSH nativa completa.
- **Memoria de reconexión:** Al conectar con un servidor, se registra automáticamente como el último conectado para su uso posterior desde la CLI (`--last` o `-`).
- **Formularios integrados:**
  - Añade (`a`) o edita (`e`) servidores mediante formularios interactivos con validación de puertos y campos obligatorios.
  - Navegación ágil con `tab`, `shift+tab`, flechas arriba/abajo y `enter`.
- **Eliminación protegida:** Elimina (`d`) servidores con diálogo de confirmación previa (`y`/`n`).
- **Selector de temas:** Pulsa `t` para abrir el selector dinámico de temas y personalizar la apariencia visual en caliente.

### Atajos de la TUI:
| Atajo | Acción |
| :--- | :--- |
| `[enter]` | Conectar al host seleccionado por SSH |
| `[/]` | Entrar en modo búsqueda / filtro |
| `[a]` | Añadir un nuevo servidor |
| `[e]` | Editar el servidor seleccionado |
| `[d]` | Eliminar el servidor seleccionado (solicita confirmación `y`/`n`) |
| `[t]` | Abrir el selector de temas visuales |
| `[esc]` | Cancelar búsqueda / cancelar formulario / volver a la lista |
| `[q]` / `[ctrl+c]` | Salir de la aplicación |

---

## Funcionalidades en Modo CLI (Línea de Comandos)

SSHPro permite ejecutar operaciones directas desde la terminal sin necesidad de abrir la TUI:

### 1. Conexión rápida por nombre
Conéctate pasando parte del nombre del servidor:
```bash
./sshpro grafana
./sshpro server-grafana
```
- Búsqueda insensible a mayúsculas/minúsculas (*case-insensitive*).
- Los guiones `-` equivalen a espacios.
- Si hay coincidencia única, conecta automáticamente.
- Si hay múltiples coincidencias, lista los nombres encontrados para que elijas con mayor precisión.

### 2. Ejecución de comandos remotos (`-c`, `--command`)
Ejecuta un comando en el servidor y sale inmediatamente sin abrir una sesión completa de shell:
```bash
./sshpro web -c "ls -la /var/www/html"
./sshpro grafana --command "systemctl status grafana-server"
```

### 3. Túnel local de puertos (`-pt`, `--port-tunnel`)
Crea un túnel local hacia el servidor remoto (*Local Port Forwarding*, `ssh -L`):
```bash
# Mapea localhost:8080 hacia localhost:80 en el servidor remoto
./sshpro web -pt 8080:80

# Mismo puerto en local y remoto
./sshpro web --port-tunnel 3000

# Túnel hacia un host interno accesible desde el servidor
./sshpro web -pt 8080:database.internal:5432

# Se puede combinar con -c para ejecutar un comando manteniendo el túnel
./sshpro web -pt 8080:80 -c "curl localhost:80"
```

### 4. Túnel inverso de puertos (`-rpt`, `--reverse-tunnel`)
Expone un servicio o puerto de tu máquina local en el servidor remoto (*Reverse Port Forwarding*, `ssh -R`):
```bash
# Expone localhost:80 de tu máquina en el puerto 8080 del servidor remoto
./sshpro web -rpt 8080:80

# Mismo puerto en ambos extremos
./sshpro web --reverse-tunnel 3000

# Combinado con comando remoto
./sshpro web -rpt 8080:3000 -c "curl localhost:8080"
```

### 5. Transferencia segura de archivos (`-scp`, `--secure-copy`)
Transfiere archivos y carpetas con `scp` usando la configuración guardada del servidor (IP, usuario, puerto y clave):
```bash
# Subir un archivo al servidor (destino con ':')
./sshpro web -scp ./archivo.txt :/var/www/html/

# Subir un archivo (por defecto el último argumento es el destino remoto)
./sshpro web -scp ./archivo.txt /var/www/html/

# Subir carpetas completas de forma recursiva con -r
./sshpro web -scp -r ./dist :/var/www/html/

# Descargar un archivo desde el servidor a la máquina local
./sshpro web -scp :/var/log/nginx/access.log ./access.log
```

### 6. Envío de clave pública (`--copy-id`)
Copia tu clave pública SSH a `~/.ssh/authorized_keys` del servidor remoto de forma nativa (compatible con Windows, Linux y macOS sin requerir scripts bash externos):
```bash
# Copia la clave asociada al host o la clave default (~/.ssh/id_ed25519.pub, id_rsa.pub)
./sshpro web --copy-id

# Copia una clave pública específica
./sshpro web --copy-id ~/.ssh/mi_llave_personal.pub
```

### 7. Chequeo de salud y latencia (`--check`, `--check-all`)
Comprueba rápidamente por TCP si el puerto SSH está accesible y mide la latencia de respuesta:
```bash
# Comprobar un servidor específico
./sshpro web --check

# Comprobar todos los servidores configurados en paralelo (concurrente)
./sshpro --check-all
```

### 8. Listado rápido en terminal (`-l`, `--list`)
Muestra una tabla con todos los servidores configurados sin abrir la TUI:
```bash
# Listar todos los servidores
./sshpro -l
./sshpro --list

# Filtrar listado por nombre
./sshpro -l web
```

### 9. Reconexión inmediata al último servidor (`--last` o `-`)
Reconéctate al último servidor con el que te conectaste (desde la TUI o CLI):
```bash
# Reconectar de inmediato
./sshpro --last
./sshpro -

# Ejecutar comando en el último servidor
./sshpro - -c "uptime"

# Abrir túnel en el último servidor
./sshpro - -pt 8080:80
```

### 10. Ayuda en línea (`-h`, `--help`)
Muestra el resumen de sintaxis general o la ayuda detallada con ejemplos para cualquier flag específica:
```bash
# Ayuda general
./sshpro -h
./sshpro --help

# Ayuda detallada de una flag específica
./sshpro --help -c
./sshpro -h -rpt
./sshpro --help -pt
./sshpro -h -scp
./sshpro --help --copy-id
./sshpro -h --check
./sshpro --help --check-all
./sshpro -h -l
./sshpro --help --last
```

---

## Configuración (JSON)

### Hosts
Ruta por defecto:
- `~/.sshpro/configured_hosts.json`

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
- `key` vacío implica usar la llave por defecto del usuario (`~/.ssh/id_rsa` o `~/.ssh/id_ed25519`). Para el caso de **servidores que están protegidos por contraseña** y no tienen una llave configurada para acceder, también se debe dejar vacío.

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

---

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

---

## Licencia
MIT — ver [`LICENCE.md`](LICENCE.md).

---

## Disclaimer
Este proyecto es una herramienta personal y no se garantiza compatibilidad con todos los entornos o configuraciones.
