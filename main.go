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
	// --- CONFIGURACIÓN (¡Cambia esto por tus datos!) ---
	host := "10.50.100.81:31022"
	usuario := "ermunoz"
	password := "LQKZgq0r$2"
	// --------------------------------------------------

	// Leemos el comando que el usuario escribe en la terminal local
	// Ejemplo: ./wrapper "LIST VOC"
	if len(os.Args) < 2 {
		fmt.Println("❌ Error: Debes enviar un comando.")
		fmt.Println("💡 Uso: ./wrapper \"LIST VOC\"")
		os.Exit(1)
	}
	comandoUsuario := os.Args[1]

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
	}
	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
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
				// Esto imprimirá en tu pantalla EXACTAMENTE lo que manda el servidor
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

	// --- FASE DE NAVEGACIÓN ---

	fmt.Println("⏳ Esperando el prompt de Linux...")
	if err := esperarHasta("$ "); err != nil {
		log.Fatalf("No encontré el prompt de Linux: %v", err)
	}

	fmt.Println("🖥️ ¡Estamos en Bash! Llamando al entorno...")
	// Aquí asumo que dejaste el comando que disparó el menú (ej. "uv\r" o tu script de perfil)
	stdin.Write([]byte("stty erase ^H && source ~/.bash_profile && uv\r"))

	fmt.Println("⏳ Esperando el menú de SISE...")
	// Le decimos a Go que atrape el menú exacto que vimos en el debug
	if err := esperarHasta("Elija la Opcion"); err != nil {
		log.Fatalf("No encontré el menú de SISE: %v", err)
	}

	fmt.Println("👉 Seleccionando opción 3 (PERU)...")
	// Enviamos el número 1 y un Enter
	stdin.Write([]byte("3\r"))
	// Bucle inteligente que reacciona a la pantalla (Versión 3.0)
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			pantalla := ansiRegex.ReplaceAllString(string(buf[:n]), "")

			if strings.TrimSpace(pantalla) != "" {
				// Puedes comentar esta línea después si no quieres ver el log en tu terminal final
				fmt.Printf("[LEYENDO]: %s\n", pantalla)
			}

			// REACCIÓN 1: Sesión colgada
			if strings.Contains(pantalla, "liberar? (S/N)") {
				fmt.Println("⚠️ Liberando sesión colgada...")
				stdin.Write([]byte("S\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			// REACCIÓN 2: Pide Usuario
			if strings.Contains(pantalla, "Codigo de Usuario") {
				fmt.Println("👉 Enviando Usuario...")
				stdin.Write([]byte("ERMUNOZ\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			// REACCIÓN 3: Pide Contraseña
			if strings.Contains(pantalla, "Password") {
				fmt.Println("👉 Enviando Contraseña...")
				passLimpio := strings.TrimSpace("aRCANGEL3#")
				fmt.Printf("[DEBUG INTERNO] Se enviará exactamente: %q\n", passLimpio)
				stdin.Write([]byte(passLimpio + "\r"))
				time.Sleep(1 * time.Second)
				continue
			}

			// REACCIÓN 4: ¡MENÚ PRINCIPAL! (El escape hacia la consola)
			if strings.Contains(pantalla, "(F) para salir") || strings.Contains(pantalla, "Seleccione la Opcion") {
				fmt.Println("✅ ¡Menú alcanzado! Saliendo con 'F' hacia el prompt TCL...")
				stdin.Write([]byte("F\r"))
				time.Sleep(1 * time.Second)
				break // ¡Rompemos el bucle! Ya terminamos el login y la navegación.
			}
		}
	}
	// Éxito B: Llegamos al menú principal de la App.
	// ¡Descomenta y cambia "MENU PRINCIPAL" por la palabra real que sale cuando entras con éxito!
	/*
		if strings.Contains(pantalla, "MENU PRINCIPAL") {
			fmt.Println("👉 Saliendo del Menú hacia la consola...")
			stdin.Write([]byte("QUIT\r")) // Comando para salir
			break // Rompemos el bucle para ir a la fase final
		}
	*/
	// --- FASE FINAL: SALIR AL PROMPT '>' ---
	// --- FASE FINAL: LA CONSOLA DE UNIVERSE ---

	fmt.Println("⏳ Esperando el prompt '>'...")
	if err := esperarHasta(">"); err != nil {
		log.Fatalf("Nunca llegué al prompt: %v", err)
	}
	fmt.Println("⏳ Logueado con éxito. Esperando el Menú Principal...")

	// 🛑 AQUÍ NECESITAMOS TU AYUDA (Lee la explicación abajo)
	time.Sleep(2 * time.Second)

	fmt.Println("👉 Saliendo de QUALITAS hacia el prompt TCL...")
	// Mandamos el comando para salir de la aplicación (AJÚSTALO)
	stdin.Write([]byte("QUIT\r"))

	fmt.Println("⏳ Esperando por fin el prompt '>'...")
	if err := esperarHasta(">"); err != nil {
		log.Fatalf("No encontré el prompt: %v", err)
	}
	// --- FASE DE EJECUCIÓN ---

	fmt.Printf("🚀 Ejecutando: %s\n", comandoUsuario)
	stdin.Write([]byte(comandoUsuario + "\r"))

	// Capturamos la respuesta hasta que vuelva a salir el prompt ">"
	var capturaFinal strings.Builder
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			capturaFinal.WriteString(chunk)

			// Quitamos el debug cuando ya estemos seguros de que funciona
			// fmt.Printf("\n[DEBUG] %q\n", chunk)

			if strings.Contains(chunk, ">") {
				break
			}
		}
		if err != nil {
			break
		}
	}

	// Limpiamos la salida
	textoLimpio := capturaFinal.String()
	textoLimpio = strings.Replace(textoLimpio, comandoUsuario+"\r\n", "", 1)
	textoLimpio = strings.TrimSuffix(textoLimpio, ">")
	textoLimpio = strings.TrimSpace(textoLimpio)

	// Salimos de UniVerse
	stdin.Write([]byte("QUIT\r"))
	// Salimos de Bash
	stdin.Write([]byte("exit\r"))

	fmt.Println("\n================ RESPUESTA DE UNIVERSE ================")
	fmt.Println(textoLimpio)
	fmt.Println("=======================================================")
}
