package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Servidor struct {
	host         string
	port         int
	user         string
	password     string
	pastaRemota  string
	pastaDestino string
}

func mensagemBoasVindas() {
	fmt.Println("=============================================")
	fmt.Println("         Bem-vindo ao Sistema FTP-GO!        ")
	fmt.Println("=============================================")
	fmt.Println("Você poderá configurar seus servidores abaixo.")
	fmt.Println("=============================================")
	fmt.Println("")
}

func lerEntrada(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(msg)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func configurarServidores() []Servidor {
	var servidores []Servidor

	qtdStr := lerEntrada("Quantos servidores deseja configurar? ")
	qtd, _ := strconv.Atoi(qtdStr)

	for i := 0; i < qtd; i++ {
		fmt.Printf("\n--- Configuração do servidor %d ---\n", i+1)

		host := lerEntrada("Host: ")
		portStr := lerEntrada("Porta (ex: 22): ")
		port, _ := strconv.Atoi(portStr)
		user := lerEntrada("Usuário: ")
		password := lerEntrada("Senha: ")
		pastaRemota := lerEntrada("Pasta remota (ex: /var/lib/vz/dump): ")
		pastaDestino := lerEntrada("Pasta local de destino: ")

		servidores = append(servidores, Servidor{
			host:         host,
			port:         port,
			user:         user,
			password:     password,
			pastaRemota:  pastaRemota,
			pastaDestino: pastaDestino,
		})
	}
	return servidores
}

func Backup(servidores []Servidor) {
	var data = time.Now()
	var nomepasta_ = "\\" + data.Format("02-01-2006")

	for i, server := range servidores {
		endereco := fmt.Sprintf("%s:%d", server.host, server.port)
		config := ssh.ClientConfig{
			User:            server.user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.Password(server.password),
			},
			Timeout: 10 * time.Second,
		}

		conexao, err := ssh.Dial("tcp", endereco, &config)
		if err != nil {
			fmt.Printf("[Servidor %d] Erro ao conectar: %v\n", i+1, err)
			continue
		}
		defer conexao.Close()

		sftp_conexao, err := sftp.NewClient(conexao)
		if err != nil {
			fmt.Printf("[Servidor %d] Erro ao iniciar SFTP: %v\n", i+1, err)
			continue
		}
		defer sftp_conexao.Close()

		arquivos, err := sftp_conexao.ReadDir(server.pastaRemota)
		if err != nil {
			fmt.Printf("[Servidor %d] Não foi possível ler a pasta remota %s: %v\n", i+1, server.pastaRemota, err)
			continue
		}

		for _, arq := range arquivos {
			nomeArquivo := arq.Name()
			if strings.HasSuffix(nomeArquivo, ".tmp") {
				continue
			}

			repositorioRemoto := path.Join(server.pastaRemota, nomeArquivo)
			arquivoRemoto, err := sftp_conexao.Open(repositorioRemoto)
			if err != nil {
				fmt.Printf("[Servidor %d] Erro ao abrir arquivo remoto %s: %v\n", i+1, nomeArquivo, err)
				return
			}

			destinoFinal := filepath.Join(server.pastaDestino, nomepasta_)
			if err := os.MkdirAll(destinoFinal, 0777); err != nil {
				fmt.Printf("[Servidor %d] Erro ao criar pasta %s: %v\n", i+1, destinoFinal, err)
				return
			}

			caminhoDestino := filepath.Join(destinoFinal, nomeArquivo)
			arquivoLocal, err := os.Create(caminhoDestino)
			if err != nil {
				fmt.Printf("[Servidor %d] Erro ao criar arquivo local %s: %v\n", i+1, caminhoDestino, err)
				return
			}
			defer arquivoLocal.Close()

			bytesCopiados, err := io.Copy(arquivoLocal, arquivoRemoto)
			if err != nil {
				fmt.Printf("[Servidor %d] Erro ao copiar arquivo %s: %v\n", i+1, nomeArquivo, err)
				return
			}
			fmt.Printf("[Servidor %d] %d bytes copiados -> %s\n", i+1, bytesCopiados, caminhoDestino)
		}
		fmt.Printf("[Servidor %d] Backup concluído com sucesso em %s\n\n", i+1, server.pastaDestino)
	}
	fmt.Println("✅ Todos os backups finalizados em " + time.Now().Format("02-01-2006"))
}

func main() {
	mensagemBoasVindas()
	servidores := configurarServidores()

	fmt.Println("\nResumo da configuração:")
	for i, s := range servidores {
		fmt.Printf("Servidor %d -> %s:%d, Usuário: %s\n", i+1, s.host, s.port, s.user)
		fmt.Printf("Pasta remota: %s\n", s.pastaRemota)
		fmt.Printf("Pasta local:  %s\n", s.pastaDestino)
		fmt.Println("---------------------------------------------")
	}

	agendador := gocron.NewScheduler(time.Local)
	agendador.Every(4).Days().At("23:00").Do(Backup, servidores)

	fmt.Println("⏳ Backup agendado para rodar a cada 4 dias às 23:00")
	fmt.Println("▶️ Deseja rodar um backup agora também? (s/n)")

	resp := lerEntrada("> ")
	if strings.ToLower(resp) == "s" {
		Backup(servidores)
	}
	agendador.StartBlocking()
}
