package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	host := "10.50.100.81:31022"
	usuario := "ermunoz"
	//usuario := "ezurita"
	password := "LQKZgq0r$2"
	//password := "ERC$51aecw$518"
	// --------------------------------------------------

	if len(os.Args) < 2 {
		fmt.Println("❌ Error: Debes enviar un comando.")
		fmt.Println("💡 Uso: ./wrapper \"LIST VOC\"")
		os.Exit(1)
	}

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
				chunk := string(buf[:n])

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
			pantalla := ansiRegex.ReplaceAllString(string(buf[:n]), "")

			if strings.TrimSpace(pantalla) != "" {
				fmt.Printf("[LEYENDO]: %s\n", pantalla)
			}

			if strings.Contains(pantalla, "liberar? (S/N)") {
				fmt.Println("⚠️ Liberando sesión colgada...")
				stdin.Write([]byte("N\r"))
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

	fmt.Println("✅ ¡Prompt '>' alcanzado! Listo para ejecutar comandos.")

	// --- FASE DE EJECUCIÓN (USANDO PSICOLOGÍA INVERSA) ---
	comandoCompleto := strings.Join(os.Args[1:], " ")

	// 1. Lo que deseamos que se ejecute (en MAYÚSCULAS)
	comandoDeseado := strings.ToUpper(comandoCompleto)

	// 2. Lo que mandamos (en MINÚSCULAS) para que UniVerse lo invierta
	comandoInvertido := strings.ToLower(comandoCompleto)

	fmt.Printf("🚀 Ejecutando (enviado invertido): %s\n", comandoDeseado)

	// Inyectamos el comando en minúsculas
	stdin.Write([]byte(comandoInvertido + "\r"))

	// Capturamos la respuesta hasta que vuelva a salir el prompt ">"
	var capturaFinal strings.Builder
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			capturaFinal.WriteString(chunk)
			textoActual := capturaFinal.String()

			textoEvaluacion := ansiRegex.ReplaceAllString(textoActual, "")
			textoEvaluacion = strings.TrimRight(textoEvaluacion, " \r\n\t")

			if strings.HasSuffix(textoEvaluacion, ">") {
				break
			}
		}
		if err != nil {
			break
		}
	}

	// Limpiamos la salida
	textoLimpio := capturaFinal.String()
	textoLimpio = ansiRegex.ReplaceAllString(textoLimpio, "")

	// OJO: Como UniVerse lo invirtió, nos va a regresar el texto en MAYÚSCULAS.
	// Por eso le decimos a Go que borre 'comandoDeseado' y no el invertido.
	textoLimpio = strings.TrimPrefix(textoLimpio, comandoDeseado+"\r\n")
	textoLimpio = strings.TrimPrefix(textoLimpio, comandoDeseado+"\r")
	textoLimpio = strings.TrimPrefix(textoLimpio, comandoDeseado)
	textoLimpio = strings.TrimSuffix(textoLimpio, ">")
	textoLimpio = strings.TrimSpace(textoLimpio)

	// Salimos de UniVerse y Bash
	stdin.Write([]byte("QUIT\r"))
	time.Sleep(500 * time.Millisecond) // Pequeña pausa para que alcance a cerrar
	stdin.Write([]byte("exit\r"))

	fmt.Println("\n================ RESPUESTA DE UNIVERSE ================")
	fmt.Println(textoLimpio)
	fmt.Println("=======================================================")
}
