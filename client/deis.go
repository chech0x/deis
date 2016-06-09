package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/deis/deis/client/parser"
	docopt "github.com/docopt/docopt-go"
)

// main exits with the return value of Command(os.Args[1:]), deferring all logic to
// a func we can test.
func main() {
	os.Exit(Command(os.Args[1:]))
}

// Command routes talka commands to their proper parser.
func Command(argv []string) int {
	usage := `
El cliente Talka para linea de comando envia llamados API hacia el controlador Talka.

Usage: talka <comando> [<args>...]

Banderas opcionales ::

  -h --help     muestra esta informacion de ayuda
  -v --version  muestra la version del cliente

comandos de autenticacion (Auth) ::

  register      registra un nuevo usuario en un controlador de talka
  login         inicia una sesion al controlador de talka
  logout        cierra la sesion activa hacia el controlador de talka

Subcomandos, use 'talka help [subcommando]' para aprender mas::

  apps          administra las aplicaciones usadas para proveer servicios
  ps            administra procesos dentro de un container de aplicacion
  config        administra las variables de ambiente que definen la configuracion de una aplicacion
  domains       administra y asigna nombres de dominio a tus aplicaciones
  builds        administra los builds que han sido creado usando 'git push'
  limits        administra los limites de recursos para tu aplicacion
  tags          administra tags para tus containers de aplicaciones
  releases      administra versiones (release) de una aplicacion
  certs         administra puntos SSL para una aplicacion

  keys          administra claves ssh usada para las instalaciones/builds con 'git push'
  perms         administra los permisos de las aplicaciones
  git           administra git para las aplicaciones
  users         administra usuarios
  version       muestra la version del cliente

Atajos de comandos, use 'talka shortcuts' para verlos todos::

  create        crea una nueva aplicacion
  scale         ayuda a escalar los procesos por tipo (web=2, worker=1)
  info          visualiza informacion sobre la aplicacion actual
  open          abre la url de la aplicacion actual en el browser predeterminado
  logs          visualiza logs de la aplicacion de forma acomulativa
  run           ejecuta un comando dentro de un container efimero de la aplicacion
  destroy       destruye una aplicacion y sus artefactos (releases,git,containers)
  pull          importa una imagen docker y la instala como un nuevo release

Use 'git push talka master' para instalar una aplicacion.
`
	// Reorganize some command line flags and commands.
	command, argv := parseArgs(argv)
	// Give docopt an optional final false arg so it doesn't call os.Exit().
	_, err := docopt.Parse(usage, []string{command}, false, "", true, false)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(argv) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: talka <comando> [<args>...]")
		return 1
	}

	// Dispatch the command, passing the argv through so subcommands can
	// re-parse it according to their usage strings.
	switch command {
	case "auth":
		err = parser.Auth(argv)
	case "ps":
		err = parser.Ps(argv)
	case "apps":
		err = parser.Apps(argv)
	case "config":
		err = parser.Config(argv)
	case "domains":
		err = parser.Domains(argv)
	case "builds":
		err = parser.Builds(argv)
	case "limits":
		err = parser.Limits(argv)
	case "tags":
		err = parser.Tags(argv)
	case "releases":
		err = parser.Releases(argv)
	case "certs":
		err = parser.Certs(argv)
	case "keys":
		err = parser.Keys(argv)
	case "perms":
		err = parser.Perms(argv)
	case "git":
		err = parser.Git(argv)
	case "users":
		err = parser.Users(argv)
	case "version":
		err = parser.Version(argv)
	case "help":
		fmt.Print(usage)
		return 0
	default:
		env := os.Environ()
		extCmd := "talka-" + command

		binary, err := exec.LookPath(extCmd)
		if err != nil {
			parser.PrintUsage()
			return 1
		}

		cmdArgv := []string{extCmd}

		cmdSplit := strings.Split(argv[0], command+":")

		if len(cmdSplit) > 1 {
			argv[0] = cmdSplit[1]
		}

		cmdArgv = append(cmdArgv, argv...)

		err = syscall.Exec(binary, cmdArgv, env)
		if err != nil {
			parser.PrintUsage()
			return 1
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// parseArgs returns the provided args with "--help" as the last arg if need be,
// expands shortcuts and formats commands to be properly routed.
func parseArgs(argv []string) (string, []string) {
	if len(argv) == 1 {
		if argv[0] == "--help" || argv[0] == "-h" {
			// rearrange "talka --help" as "talka help"
			argv[0] = "help"
		} else if argv[0] == "--version" || argv[0] == "-v" {
			// rearrange "talka --version" as "talka version"
			argv[0] = "version"
		}
	}

	if len(argv) >= 2 {
		// Rearrange "talka help <command>" to "talka <command> --help".
		if argv[0] == "help" || argv[0] == "--help" || argv[0] == "-h" {
			argv = append(argv[1:], "--help")
		}
	}

	if len(argv) > 0 {
		argv[0] = replaceShortcut(argv[0])

		index := strings.Index(argv[0], ":")

		if index != -1 {
			command := argv[0]
			return command[:index], argv
		}

		return argv[0], argv
	}

	return "", argv
}

func replaceShortcut(command string) string {
	shortcuts := map[string]string{
		"create":         "apps:create",
		"destroy":        "apps:destroy",
		"info":           "apps:info",
		"login":          "auth:login",
		"logout":         "auth:logout",
		"logs":           "apps:logs",
		"open":           "apps:open",
		"passwd":         "auth:passwd",
		"pull":           "builds:create",
		"register":       "auth:register",
		"rollback":       "releases:rollback",
		"run":            "apps:run",
		"scale":          "ps:scale",
		"sharing":        "perms:list",
		"sharing:list":   "perms:list",
		"sharing:add":    "perms:create",
		"sharing:remove": "perms:delete",
		"whoami":         "auth:whoami",
	}

	expandedCommand := shortcuts[command]
	if expandedCommand == "" {
		return command
	}

	return expandedCommand
}
