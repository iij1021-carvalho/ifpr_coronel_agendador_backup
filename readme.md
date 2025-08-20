# 🛰️ Backup de Máquinas Virtuais via SFTP (Proxmox)

Este projeto em Go conecta-se a múltiplos servidores **Proxmox via SFTP**, lista os arquivos de backup localizados no diretório `/var/lib/vz/dump`, e realiza o download automático para um diretório local, organizado por data e por servidor.

---

## 📦 Funcionalidade

- Conecta-se via SFTP a múltiplos servidores com autenticação por senha.
- Lista os arquivos de backup de VMs (por padrão, `.vma`, `.tar`, `.zst`, etc.).
- Realiza o download dos arquivos para um caminho local específico.
- Cria pastas com a **data atual (DD-MM-YYYY)** para organizar os backups.
- Exibe logs no console com o progresso do download.

---

## 🧠 Como funciona

1. A função `AdicionaServidores()` retorna a lista de servidores com:
   - Endereço IP/host
   - Porta
   - Usuário
   - Senha

2. Para cada servidor:
   - É feita uma conexão SSH/SFTP.
   - Lê os arquivos do diretório `/var/lib/vz/dump`.
   - Cria uma pasta local com o nome do servidor e a data atual.
   - Copia todos os arquivos de backup para essa pasta.

3. As pastas locais são definidas na função `RetornaLocalPasta(i)`:
exemplo:
   - `pv100-t`
   - `pv200-r`
   - `pv300-h`

4. Controlar as datas aonde serão feitos os backups:
os backups são feitos por exemplo está definido a fazer o backup a cada 4 dias.
---

## 🗂️ Estrutura das Pastas

```text
C:\
└── Backup\
    └── VMs_OVFs\
        └── Proxmox\
            ├── pv100-t\
            │   └── 19-08-2025\
            ├── pv200-r\
            │   └── 19-08-2025\
            └── pv300-h\
                └── 19-08-2025\
