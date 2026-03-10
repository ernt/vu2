package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/peterh/liner"

	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/charmap"
)

func main() {
	host := "10.50.100.81:31022"
	//host := "11.50.0.45:22"
	usuario := "ermunoz"
	//usuario := "ezurita"
	password := "LQKZgq0r$2"
	//password := "ERC$51aecw$518"
	// --------------------------------------------------
	fmt.Printf("🔌 Conectando a %s...\n", host)
	// 1. Configuración de SSH interactivo
	authInteractive := ssh.KeyboardInteractive(
		func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = password
			}
			return answers, nil
		},
	)

	config := &ssh.ClientConfig{
		User: usuario,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			authInteractive,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatalf("Fallo conexión SSH: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Fallo crear sesión: %v", err)
	}
	defer session.Close()

	// 2. Pedimos Terminal (PTY)
	modes := ssh.TerminalModes{
		ssh.ECHO:          0, // No repetir caracteres
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.VERASE:        127,
		ssh.VKILL:         21,
		ssh.IUCLC:         0,
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("Fallo pedir PTY: %v", err)
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()

	if err := session.Start("/bin/bash"); err != nil {
		log.Fatalf("Fallo iniciar shell: %v", err)
	}

	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x00|\a`)
	// 3. Función inteligente para leer hasta encontrar un texto
	buf := make([]byte, 1024)
	esperarHasta := func(marcador string) error {
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				bytesDecodificados, _ := charmap.ISO8859_1.NewDecoder().Bytes(buf[:n])
				chunk := string(bytesDecodificados)

				// --- LÍNEA MÁGICA DE DEPURACIÓN ---
				chunkLimpio := ansiRegex.ReplaceAllString(chunk, "")
				fmt.Printf("\n[DEBUG] Recibido del servidor: %q\n", chunkLimpio)
				// ----------------------------------

				if strings.Contains(chunk, marcador) {
					return nil
				}
			}
			if err != nil {
				return err
			}
		}
	}

	// --- FASE DE NAVEGACIÓN ---
	fmt.Println("⏳ Esperando el prompt de Linux...")
	if err := esperarHasta("$ "); err != nil {
		log.Fatalf("No encontré el prompt de Linux: %v", err)
	}

	fmt.Println("🖥️ ¡Estamos en Bash! Llamando al entorno...")
	stdin.Write([]byte("source ~/.bash_profile && stty erase ^H -iuclc && uv\r"))

	fmt.Println("⏳ Esperando el menú de SISE...")
	if err := esperarHasta("Elija la Opcion"); err != nil {
		log.Fatalf("No encontré el menú de SISE: %v", err)
	}

	fmt.Println("👉 Seleccionando opción 3 (PERU)...")
	stdin.Write([]byte("3\r"))

	// Bucle inteligente que reacciona a la pantalla
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			bytesDecodificados, _ := charmap.ISO8859_1.NewDecoder().Bytes(buf[:n])
			pantalla := ansiRegex.ReplaceAllString(string(bytesDecodificados), "")

			if strings.TrimSpace(pantalla) != "" {
				fmt.Printf("[LEYENDO]: %s\n", pantalla)
			}

			if strings.Contains(pantalla, "liberar? (S/N)") {
				fmt.Println("⚠️ Liberando sesión colgada...")
				stdin.Write([]byte("S\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			if strings.Contains(pantalla, "SESION a Liberar") {
				fmt.Println("🗡️ Matando la sesión colgada (Línea 1)...")
				stdin.Write([]byte("1\r"))
				time.Sleep(1 * time.Second)
				stdin.Write([]byte("F\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			if strings.Contains(pantalla, "Password") {
				fmt.Println("👉 Enviando Contraseña...")
				passLimpio := "Arcangel3#"
				fmt.Printf("[DEBUG INTERNO] Se enviará exactamente: %q\n", passLimpio)
				stdin.Write([]byte(passLimpio + "\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			if strings.Contains(pantalla, "Codigo de Usuario") {
				fmt.Println("👉 Enviando Usuario...")
				stdin.Write([]byte("ERMUNOZ\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			if strings.Contains(pantalla, "(F) para salir") || strings.Contains(pantalla, "Seleccione la Opcion") {
				fmt.Println("✅ ¡Menú alcanzado! Saliendo con 'F' hacia el prompt TCL...")
				stdin.Write([]byte("F\r"))
				time.Sleep(1 * time.Second)
				break
			}

			if strings.Contains(pantalla, "No existe la opcion") {
				fmt.Println("Pasando session")
				stdin.Write([]byte("\n\r"))
				time.Sleep(1 * time.Second)
				break
			}
		}
	}

	fmt.Println("⏳ Esperando el prompt '>'...")
	if err := esperarHasta(">"); err != nil {
		log.Fatalf("Nunca llegué al prompt: %v", err)
	}

	fmt.Println("✅ ¡Prompt '>' alcanzado! Listo para interactuar.")
	fmt.Println("=======================================================")
	fmt.Println("🔥 MODO INTERACTIVO ACTIVADO. Escribe 'EXIT' para salir.")
	fmt.Println("=======================================================")

	// Creamos un lector para escuchar tu teclado local en tiempo real
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "UniVerse > ",
		HistoryFile:     "/tmp/universe_historial.txt", // 🔥 Magia: Guarda todo en tu disco duro
		InterruptPrompt: "^C",
		EOFPrompt:       "EXIT",
	})
	if err != nil {
		log.Fatalf("Fallo al iniciar el lector de teclado: %v", err)
	}
	defer rl.Close()
	stdin.Write([]byte("term ,32767\r"))

	// Esperamos a que nos devuelva el prompt para confirmar que lo aceptó
	if err := esperarHasta(">"); err != nil {
		log.Fatalf("Fallo al apagar la paginación: %v", err)
	}
	// --- BUCLE INTERACTIVO ---
	for {
		entradaUsuario, err := rl.Readline()
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("\nSaliendo por Ctrl+C...")
			} else {
				fmt.Printf("\n❌ Error de Terminal Local: %v\n", err)
				fmt.Println("💡 CONSEJO: La librería 'liner' (flechas de historial) requiere una terminal real.")
				fmt.Println("👉 No uses el botón 'Play' verde de GoLand. Abre una ventana de terminal, compila y ejecuta './vu2'")
			}
			break // Rompemos el bucle
		}
		entradaUsuario = strings.TrimSpace(entradaUsuario)

		if entradaUsuario == "" {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(entradaUsuario)) == "EXIT" {
			fmt.Println("Saliendo de UniVerse...")
			break
		}

		// 1. INTERCEPTOR DE COMANDOS MÁGICOS
		abrirEnVSCode := false
		comandoDeseado := entradaUsuario

		// Si empieza con ".code ", activamos el modo editor y limpiamos el prefijo
		editor := ".code "
		//editor := ".antigravity"
		if strings.HasPrefix(strings.ToLower(entradaUsuario), editor) {
			abrirEnVSCode = true
			// Extraemos el comando real (ej. "CT PGM2 LANZAPRIN")
			comandoDeseado = strings.TrimSpace(entradaUsuario[6:])
			fmt.Println("📦 Interceptado: Se abrirá en VS Code al terminar...")
		}

		// 2. Aplicamos el Truco Reverso (a minúsculas)
		comandoInvertido := strings.ToLower(comandoDeseado)
		comandoISO, _ := charmap.ISO8859_1.NewEncoder().String(comandoInvertido + "\r")
		stdin.Write([]byte(comandoISO))

		// 3. Capturamos la respuesta
		var capturaFinal strings.Builder
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				bytesDecodificados, _ := charmap.ISO8859_1.NewDecoder().Bytes(buf[:n])
				chunk := string(bytesDecodificados)
				fmt.Printf("[RAYOS X] %q\n", chunk)
				capturaFinal.WriteString(chunk)
				textoActual := capturaFinal.String()

				textoEvaluacion := ansiRegex.ReplaceAllString(textoActual, "")
				textoEvaluacion = strings.TrimRight(textoEvaluacion, " \r\n\t")

				if strings.HasSuffix(textoEvaluacion, ">") || strings.HasSuffix(textoEvaluacion, "::") {
					break
				}
			}
			if err != nil {
				break
			}
		}

		// 4. Limpiamos la salida
		textoLimpio := capturaFinal.String()
		textoLimpio = ansiRegex.ReplaceAllString(textoLimpio, "")

		comandoMayus := strings.ToUpper(comandoDeseado)
		textoLimpio = strings.TrimPrefix(textoLimpio, comandoMayus+"\r\n")
		textoLimpio = strings.TrimPrefix(textoLimpio, comandoMayus+"\r")
		textoLimpio = strings.TrimPrefix(textoLimpio, comandoMayus)
		textoLimpio = strings.TrimSuffix(textoLimpio, ">")
		textoLimpio = strings.TrimSpace(textoLimpio)

		// 5. ¡LA MAGIA DE LA REDIRECCIÓN LOCAL!
		if abrirEnVSCode {
			codigoLimpio := textoLimpio

			// A. Encontrar dónde empieza el código real
			// Buscamos la primera vez que aparecen 4 números seguidos de un espacio al inicio de una línea
			reInicio := regexp.MustCompile(`(?m)^\d{4} `)
			indiceInicio := reInicio.FindStringIndex(codigoLimpio)
			if indiceInicio != nil {
				// Cortamos y tiramos a la basura todo lo que haya ANTES del primer "0001 "
				codigoLimpio = codigoLimpio[indiceInicio[0]:]
			}

			// B. Reparar las líneas macheteadas (Wrap de 80 columnas)
			// Cuando UniVerse corta una palabra, mete "\r\n" + 5 espacios. Lo reemplazamos por NADA para volver a unir la palabra (ej. C + INCO = CINCO)
			reWrap := regexp.MustCompile(`\r\n {5}`)
			codigoLimpio = reWrap.ReplaceAllString(codigoLimpio, "")

			// C. Eliminar los números de línea (ej. "0001 ")
			// Usamos la misma regla de arriba, pero ahora para borrarlos en todo el texto
			codigoLimpio = reInicio.ReplaceAllString(codigoLimpio, "")

			// D. Limpiar el prompt final que pudiera quedar pegado
			codigoLimpio = strings.TrimRight(codigoLimpio, ">\r\n ")

			// Guardar y abrir
			archivoTemp, err := os.CreateTemp("", "universe_*.bas") // .bas por UniBasic
			if err != nil {
				fmt.Println("❌ Error creando archivo temporal:", err)
			} else {
				archivoTemp.WriteString(codigoLimpio)
				archivoTemp.Close()

				cmd := exec.Command("code", archivoTemp.Name())
				err = cmd.Start()
				if err != nil {
					fmt.Println("❌ Error al abrir VS Code:", err)
				} else {
					fmt.Printf("✅ Código limpio y abierto en VS Code (%s)\n", archivoTemp.Name())
				}
			}
		} else {
			// Si no usaste el comando mágico, solo imprimimos normal en la consola
			fmt.Println(textoLimpio)
		}
	}

	// --- FASE DE CIERRE ---
	// Salimos de UniVerse y Bash solo cuando escribas EXIT
	stdin.Write([]byte("QUIT\r"))
	time.Sleep(500 * time.Millisecond)
	stdin.Write([]byte("exit\r"))

	fmt.Println("🔌 Conexión finalizada limpiamente.")

}
