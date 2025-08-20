package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type servidores struct {
	host     string
	port     int
	user     string
	password string
}

func AdicionaServidores() []servidores {
	var server servidores
	var server_ []servidores

	erro := godotenv.Load("configuracao_.env")

	if erro != nil {
		fmt.Println("erro", erro.Error())
	}

	for i := 0; i <= 2; i++ { // precisa definir aqui a quantidade de registro de servidores que vai ser rodado, por exemplo aqui defini 3 -1 no arquivo .env
		var host = os.Getenv(fmt.Sprintf("SERVER%d_HOST", i))
		var port = os.Getenv(fmt.Sprintf("SERVER%d_PORT", i))
		var user = os.Getenv(fmt.Sprintf("SERVER%d_USER", i))
		var password = os.Getenv(fmt.Sprintf("SERVER%d_PASSWORD", i))

		server.host = host
		portaa, _ := strconv.Atoi(port) // erro nao tratado

		server.port = portaa
		server.user = user
		server.password = password

		server_ = append(server_, server)
	}
	return server_
}

func Backup() {
	var server = AdicionaServidores()                 // pega a estrutura de enderecos do servidores
	var data = time.Now()                             // data atual
	var nomepasta_ = "\\" + data.Format("02-01-2006") // pega a data atual e formata para padrão 00/00/0000

	for i := range server { // laço para pegar todos os servers contido no arquivo .env
		endereco := fmt.Sprintf("%s:%d", server[i].host, server[i].port)
		config := ssh.ClientConfig{
			User:            server[i].user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.Password(server[i].password),
			},
		}

		conexao, err := ssh.Dial("tcp", endereco, &config) // seta a conexao conforme o arquivo de configuração

		if err != nil {
			fmt.Println("erro ao tentar fazer a conexão", err)
			return
		}

		sftp_conexao, erro := sftp.NewClient(conexao)

		if erro != nil {
			fmt.Println("erro ao conectar ao sftp", erro)
			return
		}

		arquivo_r, err := sftp_conexao.ReadDir("/var/lib/vz/dump") // pasta definida aonde é armazenado o backup do proxmox

		if err != nil {
			fmt.Fprintf(os.Stderr, "Não é possivel ler o diretorio %v\n", err)
			return
		}

		for _, arq := range arquivo_r {
			repositorioRemoto := "/var/lib/vz/dump/" + arq.Name() // esse var/lib/vz/dump está configurado por padrão para proxmox

			var arquivoRemoto, errrr = sftp_conexao.OpenFile(repositorioRemoto, os.O_RDONLY)
			if errrr != nil {
				fmt.Println("erro ao tentar abrir arquivo remoto", errrr)
				return
			}
			defer arquivoRemoto.Close()

			if err := os.MkdirAll(RetornaLocalPasta(i)+nomepasta_, 0777); err != nil { // cria a pasta conforme setando na função RetornalocalPasta e também cria outra pasta conforme a datal atual
				fmt.Println("erro ao tentar localizar a pasta", err)
				return
			}

			caminhoDestino := filepath.Join(RetornaLocalPasta(i)+nomepasta_, arq.Name())
			arquivoLocal, err := os.Create(caminhoDestino)

			if err != nil {
				fmt.Println("erro ao tentar localizar o caminho do arquivo", err)
				return
			}
			defer arquivoLocal.Close()

			bytesCopiados, err := io.Copy(arquivoLocal, arquivoRemoto)
			if err != nil {
				fmt.Println("erro ao copiar arquivo remoto", err)
				return
			}
			fmt.Printf("%dcopiando arquivo%s\n", bytesCopiados, caminhoDestino)
		}
		fmt.Println("Arquivo copiado com sucesso")
	}
	fmt.Println("Todos os backups foram finalizados com sucesso " + time.DateTime)
}

func RetornaLocalPasta(i int) string {
	switch i {
	case 0:
		return `C:\Backup\VMs_OVFs\Proxmox\pve100-odin` // deve ser informado a pasta aonde vai ficar o backup
	case 1:
		return `C:\Backup\VMs_OVFs\Proxmox\pve200-thor`
	case 2:
		return `C:\Backup\VMs_OVFs\Proxmox\pve300-heimdall`
	default:
		return ``
	}
}

func main() {
	fmt.Println("Iniciando Backup")
	scheduler := gocron.NewScheduler(time.Local)
	scheduler.Every(4).Days().At("14:00").Do(Backup) // executa o backup a cada 4 dias as 14 horas
	scheduler.StartBlocking()
}
