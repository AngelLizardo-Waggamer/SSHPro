package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ssh-pro/config"
	"ssh-pro/ssh"
	"ssh-pro/storage"
	"ssh-pro/ui"
)

type cliArgs struct {
	Query         string
	Command       string
	PortTunnel    string
	ReverseTunnel string
	SCPArgs       []string
	CopyID        bool
	CopyIDKey     string
	List          bool
	Check         bool
	CheckAll      bool
	Last          bool
	Help          bool
	HelpTopic     string
}

func printHelp() {
	fmt.Println("SSHPro - Gestor de conexiones SSH")
	fmt.Println()
	fmt.Println("Uso:")
	fmt.Println("  sshpro                           Inicia la interfaz interactiva (TUI)")
	fmt.Println("  sshpro <servidor>                Conecta por SSH al servidor especificado")
	fmt.Println("  sshpro <servidor> -c <comando>   Ejecuta un comando en el servidor y sale")
	fmt.Println("  sshpro <servidor> -pt <puerto>   Conecta abriendo un túnel local (-L)")
	fmt.Println("  sshpro <servidor> -rpt <puerto>  Conecta abriendo un túnel inverso (-R)")
	fmt.Println("  sshpro <servidor> -scp <args...> Transfiere archivos con scp")
	fmt.Println("  sshpro <servidor> --copy-id      Copia tu clave pública al servidor")
	fmt.Println("  sshpro <servidor> --check        Verifica conectividad y latencia del servidor")
	fmt.Println("  sshpro --check-all               Verifica conectividad de todos los servidores")
	fmt.Println("  sshpro -l, --list [filtro]       Lista los servidores guardados en terminal")
	fmt.Println("  sshpro --last, -                 Reconecta al último servidor conectado")
	fmt.Println("  sshpro --help <flag>             Muestra ayuda detallada y sintaxis de una flag")
	fmt.Println()
	fmt.Println("Flags disponibles:")
	fmt.Println("  -c,   --command <comando>       Comando a ejecutar en el servidor remoto")
	fmt.Println("  -pt,  --port-tunnel <puerto>    Túnel local a remoto (ej: 8080:80)")
	fmt.Println("  -rpt, --reverse-tunnel <puerto> Túnel inverso remoto a local (ej: 8080:80)")
	fmt.Println("  -scp, --secure-copy <args...>   Transferencia segura con scp (subida o descarga)")
	fmt.Println("        --copy-id [llave.pub]     Copia tu clave pública a ~/.ssh/authorized_keys")
	fmt.Println("        --check                   Verifica el estado y latencia de un servidor")
	fmt.Println("        --check-all               Verifica el estado de todos los servidores")
	fmt.Println("  -l,   --list [filtro]           Lista los servidores configurados en formato tabla")
	fmt.Println("        --last, -                 Reconecta al último servidor conectado")
	fmt.Println("  -h,   --help [flag]             Muestra la ayuda general o de una flag específica")
}

func normalizeHelpTopic(topic string) string {
	topic = strings.TrimSpace(strings.ToLower(topic))
	topic = strings.TrimPrefix(topic, "--")
	topic = strings.TrimPrefix(topic, "-")

	switch topic {
	case "c", "command":
		return "command"
	case "pt", "port-tunnel", "porttunnel", "tunnel":
		return "port-tunnel"
	case "rpt", "reverse-tunnel", "reversetunnel", "reverse":
		return "reverse-tunnel"
	case "scp", "secure-copy", "securecopy", "copy":
		return "secure-copy"
	case "copy-id", "copyid":
		return "copy-id"
	case "check":
		return "check"
	case "check-all", "checkall":
		return "check-all"
	case "l", "list":
		return "list"
	case "last":
		return "last"
	default:
		return topic
	}
}

func printFlagHelp(rawTopic string) bool {
	topic := normalizeHelpTopic(rawTopic)

	switch topic {
	case "command":
		fmt.Println("Ayuda para flag: -c, --command")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Ejecuta un comando en el servidor remoto de forma no interactiva")
		fmt.Println("  y retorna a tu terminal local sin iniciar una sesión completa de shell.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> -c \"<comando>\"")
		fmt.Println("  sshpro <servidor> --command \"<comando>\"")
		fmt.Println("  sshpro <servidor> -c=\"<comando>\"")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro web -c \"ls -la /var/www/html\"")
		fmt.Println("  sshpro grafana --command \"systemctl status grafana-server\"")
		fmt.Println("  sshpro db -c \"uptime && free -h\"")
		fmt.Println()
		fmt.Println("Combinaciones admitidas:")
		fmt.Println("  Se puede combinar con -pt (túnel local) o -rpt (túnel inverso).")
		fmt.Println("  No es compatible con -scp ni con --copy-id.")
		return true

	case "port-tunnel":
		fmt.Println("Ayuda para flag: -pt, --port-tunnel")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Crea un túnel de reenvío de puertos local a remoto (equivalente a ssh -L).")
		fmt.Println("  Por defecto abre una sesión interactiva manteniendo el túnel abierto, o")
		fmt.Println("  se puede combinar con -c para ejecutar un comando no interactivo.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> -pt <especificación>")
		fmt.Println("  sshpro <servidor> --port-tunnel <especificación>")
		fmt.Println("  sshpro <servidor> -pt=<especificación>")
		fmt.Println()
		fmt.Println("Formatos de especificación admitidos:")
		fmt.Println("  <puerto>                   Mismo puerto en local y remoto")
		fmt.Println("                             ej: 3000           -> 3000:localhost:3000")
		fmt.Println("  <local>:<remoto>           Puerto local y puerto en servidor")
		fmt.Println("                             ej: 8080:80        -> 8080:localhost:80")
		fmt.Println("  <local>:<host>:<remoto>    Host accesible desde el servidor")
		fmt.Println("                             ej: 8080:db.int:5432 -> 8080:db.int:5432")
		fmt.Println("  <ip_bind>:<local>:<h>:<r>  IP de enlace específica en máquina local")
		fmt.Println("                             ej: 127.0.0.1:8080:localhost:80")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro web -pt 8080:80")
		fmt.Println("  sshpro backend --port-tunnel 3000")
		fmt.Println("  sshpro bastion -pt 5432:postgres.internal:5432")
		fmt.Println("  sshpro web -pt 8080:80 -c \"curl http://localhost:8080\"")
		fmt.Println()
		fmt.Println("Combinaciones admitidas:")
		fmt.Println("  Compatible con -c y -rpt. Incompatible con -scp y --copy-id.")
		return true

	case "reverse-tunnel":
		fmt.Println("Ayuda para flag: -rpt, --reverse-tunnel")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Crea un túnel de reenvío de puertos inverso (equivalente a ssh -R).")
		fmt.Println("  Expone un puerto o servicio de tu máquina local en el servidor remoto.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> -rpt <especificación>")
		fmt.Println("  sshpro <servidor> --reverse-tunnel <especificación>")
		fmt.Println("  sshpro <servidor> -rpt=<especificación>")
		fmt.Println()
		fmt.Println("Formatos de especificación admitidos:")
		fmt.Println("  <puerto>                   Mismo puerto remoto y local")
		fmt.Println("                             ej: 8080           -> 8080:localhost:8080")
		fmt.Println("  <remoto>:<local>           Puerto en servidor y puerto en local")
		fmt.Println("                             ej: 8080:3000      -> 8080:localhost:3000")
		fmt.Println("  <remoto>:<host>:<local>    Host de destino local")
		fmt.Println("                             ej: 8080:app:3000  -> 8080:app:3000")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro web -rpt 8080:3000")
		fmt.Println("  sshpro prod --reverse-tunnel 9000")
		fmt.Println("  sshpro web -rpt 8080:3000 -c \"curl http://localhost:8080\"")
		fmt.Println()
		fmt.Println("Combinaciones admitidas:")
		fmt.Println("  Compatible con -c y -pt. Incompatible con -scp y --copy-id.")
		return true

	case "secure-copy":
		fmt.Println("Ayuda para flag: -scp, --secure-copy")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Transfiere archivos o carpetas mediante scp utilizando automáticamente")
		fmt.Println("  las credenciales, puerto (-P) y clave SSH asociadas al servidor.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> -scp [flags] <origen> <destino>")
		fmt.Println("  sshpro <servidor> --secure-copy [flags] <origen> <destino>")
		fmt.Println()
		fmt.Println("Convención de rutas:")
		fmt.Println("  - Rutas remotas: inician con ':' o 'remote:' (ej: ':/var/www/').")
		fmt.Println("  - Rutas locales: rutas estándar (en Windows 'C:\\...' se detecta como local).")
		fmt.Println("  - Por defecto: si no se usa ':', el último argumento es el destino remoto.")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  # Subida de archivo (especificando ':' remoto)")
		fmt.Println("  sshpro web -scp ./archivo.txt :/var/www/html/")
		fmt.Println()
		fmt.Println("  # Subida de directorio recursivo (-r)")
		fmt.Println("  sshpro web -scp -r ./dist/ :/var/www/html/")
		fmt.Println()
		fmt.Println("  # Descarga desde el servidor hacia local")
		fmt.Println("  sshpro web -scp :/var/log/nginx/access.log ./access.log")
		fmt.Println()
		fmt.Println("Combinaciones admitidas:")
		fmt.Println("  Incompatible con -c, -pt, -rpt y --copy-id.")
		return true

	case "copy-id":
		fmt.Println("Ayuda para flag: --copy-id")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Copia tu clave pública SSH local al archivo ~/.ssh/authorized_keys del")
		fmt.Println("  servidor remoto. Es totalmente compatible con Windows, Linux y macOS")
		fmt.Println("  sin depender del script ssh-copy-id del sistema.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> --copy-id")
		fmt.Println("  sshpro <servidor> --copy-id <ruta_clave.pub>")
		fmt.Println("  sshpro <servidor> --copy-id=<ruta_clave.pub>")
		fmt.Println()
		fmt.Println("Resolución de clave pública:")
		fmt.Println("  - Si se omite la ruta: busca la clave asociada al host en su configuración,")
		fmt.Println("    o en su defecto busca en ~/.ssh/ (id_ed25519.pub, id_ecdsa.pub, id_rsa.pub).")
		fmt.Println("  - Si se especifica una ruta: lee esa clave (.pub) y la envía.")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro web --copy-id")
		fmt.Println("  sshpro servidor-nuevo --copy-id ~/.ssh/id_rsa.pub")
		return true

	case "check":
		fmt.Println("Ayuda para flag: --check")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Verifica la disponibilidad del puerto SSH del servidor mediante un socket")
		fmt.Println("  TCP y mide el tiempo de respuesta (latencia en milisegundos).")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro <servidor> --check")
		fmt.Println("  sshpro --check")
		fmt.Println()
		fmt.Println("Nota:")
		fmt.Println("  Si se ejecuta '--check' sin especificar servidor, comprobará todos")
		fmt.Println("  los servidores de la lista (equivalente a --check-all).")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro web --check")
		fmt.Println("  sshpro grafana --check")
		return true

	case "check-all":
		fmt.Println("Ayuda para flag: --check-all")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Comprueba concurrentemente (en paralelo con goroutines) la disponibilidad")
		fmt.Println("  y latencia de TODOS los servidores configurados en ~/.sshpro/configured_hosts.json.")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro --check-all")
		fmt.Println()
		fmt.Println("Salida:")
		fmt.Println("  Muestra una tabla con el estado (🟢 OK / 🔴 FAIL), servidor, dirección,")
		fmt.Println("  latencia y resumen con totales en línea e inaccesibles.")
		return true

	case "list":
		fmt.Println("Ayuda para flag: -l, --list")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Imprime en la terminal una tabla con los servidores configurados sin")
		fmt.Println("  iniciar la interfaz de usuario interactiva (TUI).")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro -l")
		fmt.Println("  sshpro --list")
		fmt.Println("  sshpro -l <filtro>")
		fmt.Println("  sshpro --list <filtro>")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro -l")
		fmt.Println("  sshpro -l web")
		fmt.Println("  sshpro --list prod")
		return true

	case "last":
		fmt.Println("Ayuda para opción: --last, -")
		fmt.Println()
		fmt.Println("Descripción:")
		fmt.Println("  Se reconecta automáticamente al último servidor con el que estableciste")
		fmt.Println("  conexión exitosa (ya sea desde la TUI o desde la línea de comandos).")
		fmt.Println()
		fmt.Println("Sintaxis:")
		fmt.Println("  sshpro --last")
		fmt.Println("  sshpro -")
		fmt.Println("  sshpro - <flags adicionales>")
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  sshpro -")
		fmt.Println("  sshpro --last")
		fmt.Println("  sshpro - -c \"uptime\"")
		fmt.Println("  sshpro - -pt 8080:80")
		fmt.Println("  sshpro - -scp ./build.zip :/var/www/")
		return true

	default:
		return false
	}
}

func parseCLIArgs(args []string) (*cliArgs, error) {
	parsed := &cliArgs{}

	// Detección de solicitud de ayuda (general o por flag específica)
	hasHelp := false
	helpTopic := ""
	for i, a := range args {
		if a == "-h" || a == "--help" || a == "help" {
			hasHelp = true
			if i+1 < len(args) && args[i+1] != "-h" && args[i+1] != "--help" && args[i+1] != "help" {
				helpTopic = args[i+1]
			}
		} else if strings.HasPrefix(a, "--help=") {
			hasHelp = true
			helpTopic = strings.TrimPrefix(a, "--help=")
		} else if strings.HasPrefix(a, "-h=") {
			hasHelp = true
			helpTopic = strings.TrimPrefix(a, "-h=")
		}
	}

	if hasHelp {
		if helpTopic == "" {
			for _, a := range args {
				if a != "-h" && a != "--help" && a != "help" && !strings.HasPrefix(a, "--help=") && !strings.HasPrefix(a, "-h=") {
					helpTopic = a
					break
				}
			}
		}
		parsed.Help = true
		parsed.HelpTopic = helpTopic
		return parsed, nil
	}

	var queryParts []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			queryParts = append(queryParts, args[i+1:]...)
			break
		}

		if arg == "-" || arg == "--last" {
			parsed.Last = true
			continue
		}

		if arg == "-l" || arg == "--list" {
			parsed.List = true
			continue
		}

		if arg == "--check-all" {
			parsed.CheckAll = true
			continue
		}

		if arg == "--check" {
			parsed.Check = true
			continue
		}

		if arg == "--copy-id" {
			parsed.CopyID = true
			if i+1 < len(args) && strings.HasSuffix(args[i+1], ".pub") {
				parsed.CopyIDKey = args[i+1]
				i++
			}
			continue
		}

		if strings.HasPrefix(arg, "--copy-id=") {
			parsed.CopyID = true
			parsed.CopyIDKey = strings.TrimPrefix(arg, "--copy-id=")
			continue
		}

		if arg == "-c" || arg == "--command" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("la flag %s requiere un argumento", arg)
			}
			if parsed.Command != "" {
				return nil, fmt.Errorf("la flag %s solo se puede especificar una vez", arg)
			}
			cmdVal := args[i+1]
			if strings.TrimSpace(cmdVal) == "" {
				return nil, fmt.Errorf("el comando especificado en %s no puede estar vacío", arg)
			}
			parsed.Command = cmdVal
			i++
			continue
		}

		if strings.HasPrefix(arg, "-c=") {
			if parsed.Command != "" {
				return nil, fmt.Errorf("la flag -c solo se puede especificar una vez")
			}
			cmdVal := strings.TrimPrefix(arg, "-c=")
			if strings.TrimSpace(cmdVal) == "" {
				return nil, fmt.Errorf("el comando especificado en -c no puede estar vacío")
			}
			parsed.Command = cmdVal
			continue
		}

		if strings.HasPrefix(arg, "--command=") {
			if parsed.Command != "" {
				return nil, fmt.Errorf("la flag --command solo se puede especificar una vez")
			}
			cmdVal := strings.TrimPrefix(arg, "--command=")
			if strings.TrimSpace(cmdVal) == "" {
				return nil, fmt.Errorf("el comando especificado en --command no puede estar vacío")
			}
			parsed.Command = cmdVal
			continue
		}

		if arg == "-pt" || arg == "--port-tunnel" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("la flag %s requiere un argumento", arg)
			}
			if parsed.PortTunnel != "" {
				return nil, fmt.Errorf("la flag %s solo se puede especificar una vez", arg)
			}
			tunnelVal := args[i+1]
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel de puerto especificado en %s no puede estar vacío", arg)
			}
			parsed.PortTunnel = tunnelVal
			i++
			continue
		}

		if strings.HasPrefix(arg, "-pt=") {
			if parsed.PortTunnel != "" {
				return nil, fmt.Errorf("la flag -pt solo se puede especificar una vez")
			}
			tunnelVal := strings.TrimPrefix(arg, "-pt=")
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel de puerto especificado en -pt no puede estar vacío")
			}
			parsed.PortTunnel = tunnelVal
			continue
		}

		if strings.HasPrefix(arg, "--port-tunnel=") {
			if parsed.PortTunnel != "" {
				return nil, fmt.Errorf("la flag --port-tunnel solo se puede especificar una vez")
			}
			tunnelVal := strings.TrimPrefix(arg, "--port-tunnel=")
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel de puerto especificado en --port-tunnel no puede estar vacío")
			}
			parsed.PortTunnel = tunnelVal
			continue
		}

		if arg == "-rpt" || arg == "--reverse-tunnel" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("la flag %s requiere un argumento", arg)
			}
			if parsed.ReverseTunnel != "" {
				return nil, fmt.Errorf("la flag %s solo se puede especificar una vez", arg)
			}
			tunnelVal := args[i+1]
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel inverso especificado en %s no puede estar vacío", arg)
			}
			parsed.ReverseTunnel = tunnelVal
			i++
			continue
		}

		if strings.HasPrefix(arg, "-rpt=") {
			if parsed.ReverseTunnel != "" {
				return nil, fmt.Errorf("la flag -rpt solo se puede especificar una vez")
			}
			tunnelVal := strings.TrimPrefix(arg, "-rpt=")
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel inverso especificado en -rpt no puede estar vacío")
			}
			parsed.ReverseTunnel = tunnelVal
			continue
		}

		if strings.HasPrefix(arg, "--reverse-tunnel=") {
			if parsed.ReverseTunnel != "" {
				return nil, fmt.Errorf("la flag --reverse-tunnel solo se puede especificar una vez")
			}
			tunnelVal := strings.TrimPrefix(arg, "--reverse-tunnel=")
			if strings.TrimSpace(tunnelVal) == "" {
				return nil, fmt.Errorf("el túnel inverso especificado en --reverse-tunnel no puede estar vacío")
			}
			parsed.ReverseTunnel = tunnelVal
			continue
		}

		if strings.HasPrefix(arg, "-scp=") || strings.HasPrefix(arg, "--secure-copy=") {
			if len(parsed.SCPArgs) > 0 {
				return nil, fmt.Errorf("la flag scp solo se puede especificar una vez")
			}
			var val string
			if strings.HasPrefix(arg, "-scp=") {
				val = strings.TrimPrefix(arg, "-scp=")
			} else {
				val = strings.TrimPrefix(arg, "--secure-copy=")
			}
			val = strings.TrimSpace(val)
			if val == "" {
				return nil, fmt.Errorf("la flag scp requiere argumentos de origen y destino")
			}
			parsed.SCPArgs = strings.Fields(val)
			continue
		}

		if arg == "-scp" || arg == "--secure-copy" {
			if len(parsed.SCPArgs) > 0 {
				return nil, fmt.Errorf("la flag %s solo se puede especificar una vez", arg)
			}
			if i+1 >= len(args) {
				return nil, fmt.Errorf("la flag %s requiere argumentos de origen y destino", arg)
			}

			if len(queryParts) > 0 {
				rest := args[i+1:]
				if len(rest) == 1 && strings.Contains(rest[0], " ") {
					parsed.SCPArgs = strings.Fields(rest[0])
				} else {
					parsed.SCPArgs = rest
				}
				i = len(args)
				continue
			}

			if strings.Contains(args[i+1], " ") {
				parsed.SCPArgs = strings.Fields(args[i+1])
				i++
				continue
			}

			rest := args[i+1:]
			if len(rest) >= 3 && !strings.HasPrefix(rest[len(rest)-1], "-") && !strings.HasPrefix(rest[len(rest)-1], ":") {
				queryParts = append(queryParts, rest[len(rest)-1])
				parsed.SCPArgs = rest[:len(rest)-1]
			} else {
				parsed.SCPArgs = rest
			}
			i = len(args)
			continue
		}

		queryParts = append(queryParts, arg)
	}

	parsed.Query = strings.TrimSpace(strings.Join(queryParts, " "))

	if len(parsed.SCPArgs) > 0 {
		if parsed.Command != "" {
			return nil, fmt.Errorf("las flags -c y -scp no se pueden usar juntas")
		}
		if parsed.PortTunnel != "" {
			return nil, fmt.Errorf("las flags -pt y -scp no se pueden usar juntas")
		}
		if parsed.ReverseTunnel != "" {
			return nil, fmt.Errorf("las flags -rpt y -scp no se pueden usar juntas")
		}
		if parsed.CopyID {
			return nil, fmt.Errorf("las flags --copy-id y -scp no se pueden usar juntas")
		}
	}

	if parsed.CopyID {
		if parsed.Command != "" {
			return nil, fmt.Errorf("las flags -c y --copy-id no se pueden usar juntas")
		}
		if parsed.PortTunnel != "" {
			return nil, fmt.Errorf("las flags -pt y --copy-id no se pueden usar juntas")
		}
		if parsed.ReverseTunnel != "" {
			return nil, fmt.Errorf("las flags -rpt y --copy-id no se pueden usar juntas")
		}
	}

	if parsed.PortTunnel != "" {
		if _, err := ssh.FormatPortTunnel(parsed.PortTunnel); err != nil {
			return nil, err
		}
	}

	if parsed.ReverseTunnel != "" {
		if _, err := ssh.FormatPortTunnel(parsed.ReverseTunnel); err != nil {
			return nil, err
		}
	}

	// Si no es List, CheckAll, Last ni Help, requiere un servidor
	if !parsed.List && !parsed.CheckAll && !parsed.Last && !parsed.Help {
		if (parsed.Command != "" || parsed.PortTunnel != "" || parsed.ReverseTunnel != "" || len(parsed.SCPArgs) > 0 || parsed.CopyID) && parsed.Query == "" {
			return nil, fmt.Errorf("debes especificar el nombre de un servidor")
		}
	}

	return parsed, nil
}

func printHostsTable(hosts []storage.Host) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NOMBRE\tUSUARIO\tDIRECCIÓN / IP\tPUERTO\tCLAVE SSH")
	fmt.Fprintln(w, "------\t-------\t--------------\t------\t---------")
	for _, h := range hosts {
		keyStr := h.Key
		if keyStr == "" {
			keyStr = "(default)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", h.Name, h.User, h.IP, h.PortOrDefault(), keyStr)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d servidores\n", len(hosts))
}

func printCheckResults(results []ssh.HostStatus) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ESTADO\tSERVIDOR\tDIRECCIÓN\tLATENCIA\tDETALLES")
	fmt.Fprintln(w, "------\t--------\t---------\t--------\t--------")
	onlineCount := 0
	for _, res := range results {
		statusStr := "🟢 OK"
		latStr := fmt.Sprintf("%dms", res.Latency.Milliseconds())
		errStr := "-"
		if !res.Online {
			statusStr = "🔴 FAIL"
			latStr = "-"
			errStr = res.Error
			if strings.Contains(errStr, "timeout") {
				errStr = "tiempo de espera agotado"
			} else if strings.Contains(errStr, "actively refused") || strings.Contains(errStr, "refused") {
				errStr = "conexión rechazada"
			}
		} else {
			onlineCount++
		}
		addr := fmt.Sprintf("%s:%d", res.Host.IP, res.Host.PortOrDefault())
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", statusStr, res.Host.Name, addr, latStr, errStr)
	}
	w.Flush()
	fmt.Printf("\nVerificados: %d | En línea: %d | Inaccesibles: %d\n", len(results), onlineCount, len(results)-onlineCount)
}

func main() {
	if len(os.Args) > 1 {
		cli, err := parseCLIArgs(os.Args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if cli.Help {
			if cli.HelpTopic != "" {
				if !printFlagHelp(cli.HelpTopic) {
					fmt.Fprintf(os.Stderr, "No se encontró ayuda específica para '%s'.\nEjecuta 'sshpro --help' para ver la lista completa de opciones.\n", cli.HelpTopic)
					os.Exit(1)
				}
				return
			}
			printHelp()
			return
		}

		path, err := storage.DefaultPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "No se pudo resolver la ruta de configuración: %v\n", err)
			os.Exit(1)
		}

		store := storage.NewStore(path)
		hosts, err := store.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "No se pudo cargar la configuración: %v\n", err)
			os.Exit(1)
		}

		// Feature: -l, --list
		if cli.List {
			targetHosts := hosts
			if cli.Query != "" {
				targetHosts = storage.FindHostsByName(hosts, cli.Query)
				if len(targetHosts) == 0 {
					fmt.Printf("No se encontró ningún servidor que coincida con '%s'\n", cli.Query)
					return
				}
			} else if len(targetHosts) == 0 {
				fmt.Println("No hay servidores configurados en la lista.")
				return
			}
			printHostsTable(targetHosts)
			return
		}

		// Feature: --check-all o --check sin query
		if cli.CheckAll || (cli.Check && cli.Query == "") {
			if len(hosts) == 0 {
				fmt.Println("No hay servidores configurados para verificar.")
				return
			}
			fmt.Printf("Verificando conectividad de %d servidor(es)...\n", len(hosts))
			results := ssh.CheckHostsConcurrently(hosts, 3*time.Second)
			printCheckResults(results)
			return
		}

		// Feature: --last o '-'
		if cli.Last {
			lastHost, err := storage.GetLastHost()
			if err != nil || lastHost == "" {
				fmt.Fprintln(os.Stderr, "Error: no hay ningún servidor registrado como último conectado")
				os.Exit(1)
			}
			cli.Query = lastHost
		}

		// Si no hay query, requerir nombre de servidor
		if cli.Query == "" {
			fmt.Fprintln(os.Stderr, "Error: debes especificar el nombre de un servidor")
			os.Exit(1)
		}

		matches := storage.FindHostsByName(hosts, cli.Query)
		if len(matches) == 0 {
			fmt.Println("No se encontró ningún servidor que coincida con el nombre especificado")
			return
		}
		if len(matches) > 1 {
			fmt.Println("Se encontraron 2 o más coincidencias de nombre:")
			for _, host := range matches {
				fmt.Printf("* %s\n", host.Name)
			}
			fmt.Println("Intenta con un nombre más específico")
			return
		}

		targetHost := matches[0]

		// Feature: --check para un host específico
		if cli.Check {
			fmt.Printf("Verificando conectividad con '%s' (%s:%d)...\n", targetHost.Name, targetHost.IP, targetHost.PortOrDefault())
			result := ssh.CheckHost(targetHost, 3*time.Second)
			printCheckResults([]ssh.HostStatus{result})
			return
		}

		// Feature: --copy-id
		if cli.CopyID {
			keyPath, pubKey, err := ssh.FindPublicKey(cli.CopyIDKey, targetHost)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error con la clave pública: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Instalando clave pública '%s' en %s...\n", keyPath, targetHost.Address())
			cmd, err := ssh.BuildCopyIDCommand(targetHost, pubKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error al preparar comando: %v\n", err)
				os.Exit(1)
			}

			if err := cmd.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					os.Exit(exitErr.ExitCode())
				}
				fmt.Fprintf(os.Stderr, "Error al copiar clave pública: %v\n", err)
				os.Exit(1)
			}
			_ = storage.SaveLastHost(targetHost.Name)
			fmt.Printf("¡Clave pública instalada con éxito en %s!\n", targetHost.Address())
			return
		}

		var cmd *exec.Cmd
		if len(cli.SCPArgs) > 0 {
			cmd, err = ssh.BuildSCPCommand(targetHost, cli.SCPArgs)
		} else {
			cmd, err = ssh.BuildCommandWithOptions(targetHost, ssh.ConnectOptions{
				PortTunnel:    cli.PortTunnel,
				ReverseTunnel: cli.ReverseTunnel,
				RemoteCommand: cli.Command,
			})
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error al preparar comando: %v\n", err)
			os.Exit(1)
		}

		_ = storage.SaveLastHost(targetHost.Name)

		if err := cmd.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error de ejecución: %v\n", err)
			os.Exit(1)
		}
		return
	}

	path, err := storage.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo resolver la ruta de configuración: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewStore(path)
	hosts, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo cargar la configuración: %v\n", err)
		os.Exit(1)
	}

	if _, err := config.EnsureThemeFile(); err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo preparar el tema: %v\n", err)
		os.Exit(1)
	}

	themeConfig, err := config.LoadThemeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo cargar el tema: %v\n", err)
		os.Exit(1)
	}

	model := ui.NewModel(store, hosts, themeConfig)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error al ejecutar la UI: %v\n", err)
		os.Exit(1)
	}
}
