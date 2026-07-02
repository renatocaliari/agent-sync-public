# CloakPipe: PII Masking para LLM Pipelines em Go

**Stack:** Rust binary (sidecar) · Proxy HTTP OpenAI-compatible · <25ms latência · MIT

---

> **🚨 ATENÇÃO — Não publiquei binários/imagens pre-built (jun/2026)**
>
> O repo [`rohansx/cloakpipe`](https://github.com/rohansx/cloakpipe/releases) tem 10 releases
> (v0.1.0 → v0.10.0), mas **TODAS têm `assets: []`** — nunca publicou binários pré-compilados.
> O pacote GHCR `ghcr.io/rohansx/cloakpipe` também não existe (404).
>
> README oficial confirma:
> > "Prebuilt binaries, `cargo install cloakpipe`, and a published image (`ghcr.io`) are on the roadmap — `docker build` or build from source for now."
>
> **Como instalar (únicas opções válidas):**
>
> | Método | Comando |
> |--------|---------|
> | Build from source (Rust) | `git clone https://github.com/rohansx/cloakpipe && cd cloakpipe && cargo build --release -p cloakpipe-cli && cp target/release/cloakpipe /usr/local/bin/` |
> | Docker compose com build | `cloakpipe: { build: { context: https://github.com/rohansx/cloakpipe.git#main } }` |
> | Docker multi-stage no seu Dockerfile | `FROM rust:1.88 AS builder ... && cargo build --release -p cloakpipe-cli` |
>
> **NÃO FAÇA ISSO (vai falhar):**
> - `curl ... | grep browser_download_url | cut -d\" -f4 | xargs curl ...` — `assets: []` retorna string vazia, `xargs curl` falha com "no URL specified"
> - `image: ghcr.io/rohansx/cloakpipe:latest` no compose — 404 no registry
> - Achar que `cargo install cloakpipe` funciona — só funciona se publicarem no crates.io (ainda não publicaram build script)

---

## Quando usar

Seu Go app envia dados para LLM (OpenAI, Anthropic, etc.) e precisa **sanitizar PII antes do envio** com:
- Substituição reversível (LLM vê `PERSON_042`, resposta é re-hidratada)
- Detecção de nomes, endereços, organizações via NER (não só regex)
- Mínima latência e zero dependência runtime

---

## Arquitetura

```
┌──────────────────────┐     ┌──────────────────┐     ┌───────────┐
│  Go App              │────▶│  CloakPipe :8900  │────▶│  LLM API  │
│                      │     │                   │     │           │
│  http.Client → :8900 │     │  regex <1ms       │     │  nunca vê │
│  (só muda base URL)  │     │  + NER 5-15ms     │     │  dado real│
│                      │     │  + vault AES-256   │     │           │
│  OPENAI_BASE_URL=    │     │  + re-hidratação   │     │           │
│  http://127.0.0.1    │     │                   │     │           │
│  :8900/v1            │     │  ＜ 25ms total     │     │           │
└──────────────────────┘     └──────────────────┘     └───────────┘
```

**Fluxo:** Detect → Mask → Proxy → Unmask (response re-hidratada automaticamente)

---

## Instalação

> **Não puxe de GitHub Releases nem de `ghcr.io/rohansx/cloakpipe`** — não existem. Veja aviso no topo.

### Tempo de build

**Primeira build: ~45 min** (workspace 8 crates + ort-sys + tokenizers + ICU + ring + lopdf).
**Com GHA cache populado: 5-10 min** (só recompila crates alteradas).

Não cancelar mid-build! Cache de GHA só popula no fim. Cancelar = próxima build começa do zero.

### Binary local (recomendado para dev)

```bash
git clone https://github.com/rohansx/cloakpipe
cd cloakpipe
cargo build --release -p cloakpipe-cli
cp target/release/cloakpipe /usr/local/bin/
```

Release binary ~15MB. Zero dependências runtime. Roda em qualquer Linux/macOS.

### Docker multi-stage

> **Não use `rust:slim`** — falta `libstdc++` (linker `cc` falha com `cannot find -lstdc++`).
> Use a imagem full (`rust:1.88-bookworm`) e instale `g++`, OU use Alpine com `musl-dev` mas atenção a outras deps nativas (`libsqlite3-sys`, `onig-sys`, `ring`, `ort`).

```dockerfile
# Recomendado — Debian full (compatível com o Dockerfile oficial do cloakpipe)
FROM rust:1.88-bookworm AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config libssl-dev ca-certificates git g++ \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /build
RUN git clone --depth 1 --branch main https://github.com/rohansx/cloakpipe.git . \
    && cargo build --release -p cloakpipe-cli
# runtime stage ...
```

---

## Configuração para Português BR

### `cloakpipe.toml` completo

```toml
[proxy]
listen = "127.0.0.1:8900"
upstream = "https://api.openai.com"
api_key_env = "OPENAI_API_KEY"
timeout_seconds = 120
max_concurrent = 256

[vault]
path = "./vault.enc"
encryption = "aes-256-gcm"
key_env = "CLOAKPIPE_VAULT_KEY"

[detection]
secrets = false
financial = false
dates = true
emails = true
phone_numbers = true
ip_addresses = true
urls_internal = false

[detection.ner]
enabled = true
backend = "distilbert_pii"
confidence_threshold = 0.85
entity_types = ["PERSON", "ORG", "LOC", "DATE"]

[detection.custom]
patterns = [
  # CPF: formatado XXX.XXX.XXX-XX ou 11 dígitos
  { name = "br_cpf",
    regex = "\\b\\d{3}\\.\\d{3}\\.\\d{3}-\\d{2}\\b|\\b\\d{11}\\b",
    category = "BR_CPF" },

  # CNPJ: formatado XX.XXX.XXX/XXXX-XX ou 14 dígitos
  { name = "br_cnpj",
    regex = "\\b\\d{2}\\.\\d{3}\\.\\d{3}/\\d{4}-\\d{2}\\b|\\b\\d{14}\\b",
    category = "BR_CNPJ" },

  # CEP: 01310-100 ou 01310100
  { name = "br_cep",
    regex = "\\b\\d{5}-\\d{3}\\b|\\b\\d{8}\\b",
    category = "BR_CEP" },

  # Telefone: +55 (11) 9XXXX-XXXX (celular) ou [2-5]XXXX-XXXX (fixo)
  { name = "br_phone",
    regex = "(?:\\+55[\\s\\-]?)?\\(?[1-9]\\d\\)?[\\s\\-]?9\\d{4}[\\s\\-]?\\d{4}|(?:\\+55[\\s\\-]?)?\\(?[1-9]\\d\\)?[\\s\\-]?[2-5]\\d{3}[\\s\\-]?\\d{4}",
    category = "BR_PHONE" },
]

[detection.overrides]
preserve = []
force = []

[audit]
enabled = true
log_path = "./audit/"
format = "jsonl"
retention_days = 90
log_entities = true
log_mappings = false
```

---

## Integração Go

### Cliente HTTP padrão (OpenAI SDK)

A única mudança no código Go é **apontar a base URL** para o CloakPipe:

```go
// ANTES
client := openai.New(os.Getenv("OPENAI_API_KEY"))

// DEPOIS — mesma SDK, só muda base URL
client := openai.New(os.Getenv("OPENAI_API_KEY"))
client.BaseURL = "http://127.0.0.1:8900/v1"
```

Ou via variável de ambiente (sem mudar código):

```bash
export OPENAI_BASE_URL=http://127.0.0.1:8900/v1
```

### HTTP client genérico (qualquer LLM)

```go
type LLMClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewLLMClient() *LLMClient {
    return &LLMClient{
        baseURL: "http://127.0.0.1:8900/v1",
        httpClient: &http.Client{
            Timeout: 120 * time.Second,
        },
    }
}

func (c *LLMClient) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/chat/completions",
        bytes.NewReader(body))
    httpReq.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("cloakpipe request: %w", err)
    }
    defer resp.Body.Close()

    var result ChatResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

### Transport proxy (qualquer requisição HTTP)

Usa `httputil.ReverseProxy` ou `http.Transport`:

```go
// Rota todo tráfego LLM via CloakPipe
transport := &http.Transport{
    Proxy: http.ProxyURL(&url.URL{
        Scheme: "http",
        Host:   "127.0.0.1:8900",
    }),
}
client := &http.Client{Transport: transport}
```

---

## Deploy como Sidecar

### systemd (Linux)

```ini
[Unit]
Description=CloakPipe PII Masking Proxy
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloakpipe --config /etc/cloakpipe.toml start
Environment=OPENAI_API_KEY=sk-...
Environment=CLOAKPIPE_VAULT_KEY=<openssl rand -hex 32>
Restart=always
User=cloakpipe
Group=cloakpipe

[Install]
WantedBy=multi-user.target
```

### docker-compose (junto com o Go app)

> **NÃO use `image: ghcr.io/rohansx/cloakpipe`** — não existe. Use `build:` com `context` apontando pro repo git.

```yaml
version: "3.9"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_BASE_URL=http://cloakpipe:8900/v1
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - cloakpipe

  cloakpipe:
    # Build from source — rohansx/cloakpipe não publica imagens pre-built
    build:
      context: https://github.com/rohansx/cloakpipe.git#main
      dockerfile: Dockerfile
    image: cloakpipe:local  # tag local (evita pull acidental do registry)
    ports:
      - "8900:8900"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - CLOAKPIPE_VAULT_KEY=${CLOAKPIPE_VAULT_KEY}
    volumes:
      - ./cloakpipe.toml:/etc/cloakpipe.toml:ro
      - cloakpipe-data:/data

volumes:
  cloakpipe-data:
```

---

## Limitações Conhecidas

| Aspecto | Detalhe |
|---------|---------|
| **CPF/CNPJ** | Regex-only (sem validação de dígito mod-11). Pode capturar falsos positivos (ex: `000.000.000-00`). Aceitável pois falso positivo é seguro vs. vazar dado real. |
| **Nome/LOC/ORG** | Dependente do modelo DistilBERT-PII (baixado na primeira execução, ~63MB). Performance depende do CPU. |
| **Idioma** | DistilBERT-PII é treinado principalmente em inglês. Nomes brasileiros podem ter recall menor que nomes ingleses. Custom regex compensa para documentos BR. |
| **Vault** | AES-256-GCM. A chave (`CLOAKPIPE_VAULT_KEY`) precisa ser secreta e persistente entre restartos — se perder a chave, tokens não são recuperáveis. |

---

## Alternativas para Go Puro (sem sidecar)

Para casos onde rodar um segundo processo é inviável, considerar:

- **taoq-ai/wuming** — Go puro, regex-based, LGPD preset. Detecta CPF, CNPJ, CEP, phone, PIS, CNH. **Não detecta nomes ou endereços** (sem NER).
- **veil-services/veil-go** — Go puro, detecta CPF, email, CC. Foco em secrets.

Se precisar de NER em Go sem sidecar: usar Presidio como REST API (Docker) + [`presidio-go-client`](https://github.com/CodeRunRepeat/presidio-go-client).

---

## Referências

- [CloakPipe GitHub](https://github.com/rohansx/cloakpipe)
- [CloakPipe Docs](https://docs.cloakpipe.co)
- [wuming BR detectors](https://github.com/taoq-ai/wuming/tree/main/adapter/detector/br) (regex originais)
- [Presidio + spaCy pt-BR](https://microsoft.github.io/presidio/analyzer/languages/)
