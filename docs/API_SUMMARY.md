# Boopateway API: resumo para integração

Documento de referência para outra IA (ou dev) entender e implementar um cliente da API
Boopateway. Tudo aqui reflete o comportamento real do servidor.

## 1. Visão geral

Gateway de pagamento com identificação facial:

1. A pessoa é **cadastrada** com nome, e-mail, foto e um código de 4 dígitos.
2. Um **pagamento** é criado enviando **valor + foto**. A API identifica a pessoa pela foto.
3. O pagamento é **confirmado** com o código de 4 dígitos da pessoa.

```
POST /persons  ──►  POST /payments  ──►  POST /payments/{id}/confirm
 (foto + code)       (valor + foto)          (code de 4 dígitos)
                          │
                ┌─────────┴──────────┐
        pessoa encontrada     não encontrada
     pending_confirmation     cancel_notfound   (estado final)
                │
       ┌────────┴─────────┐
   code correto     3 codes errados
    confirmed     cancel_invalid_code           (estados finais)
```

## 2. Regras gerais

| Item | Valor |
|---|---|
| Base URL (dev) | `http://localhost:8080` |
| Autenticação do cliente | Nenhuma (o token do parceiro fica só no servidor, em `BOOP_TOKEN`) |
| Formato de resposta | Sempre `application/json` |
| Formato de erro | `{"error": "<mensagem>"}` (alguns erros trazem campos extras, ver cada endpoint) |
| Valores monetários | Inteiro em **centavos** (`1990` = R$ 19,90). Máximo `10000000000` |
| Imagens | Campo multipart `image`, **JPEG ou PNG**, até **8 MiB**. O tipo é detectado pelos bytes, não pela extensão |
| Tamanho máximo do body | 10 MiB |
| Código da pessoa | String com exatamente 4 dígitos (`^[0-9]{4}$`). Padrão `"0000"` |
| Tentativas de confirmação | 3. Na 3ª errada a transação vira `cancel_invalid_code` |
| Datas | ISO 8601 UTC, ex.: `2026-10-17T19:37:57.639Z` |

### 2.1 Idempotência (obrigatória em todo `POST`)

Todo `POST` **exige** o header `Idempotency-Key` (string de 1 a 255 caracteres, recomendado UUID).
A chave é única **por rota** (método + path) e vale **24h**.

| Situação | Resultado |
|---|---|
| Header ausente ou com mais de 255 caracteres | `400` `{"error":"header Idempotency-Key é obrigatório (máx. 255 caracteres)"}` |
| Primeira requisição com a chave | Processa normalmente e grava a resposta (status + body) |
| Mesma chave + mesmo payload | Devolve a resposta gravada (mesmo status e body) + header `Idempotent-Replayed: true`. **Nada é executado de novo** |
| Mesma chave + payload diferente | `422` `{"error":"Idempotency-Key já usada com outro payload"}` |
| Mesma chave enquanto a 1ª ainda processa | `409` `{"error":"requisição com esta Idempotency-Key ainda em processamento"}` |
| Resposta `5xx` | A chave é liberada; pode tentar de novo com a mesma chave |

Regras práticas para o cliente:

- Gere **uma chave nova por operação de negócio** e reutilize-a **apenas em retentativas** da mesma operação (timeout, erro de rede, 5xx).
- Respostas `4xx` também ficam gravadas: reenviar a mesma chave devolve o mesmo erro. Para uma **nova tentativa de código** no confirm, use **chave nova**.
- Em multipart o "mesmo payload" é comparado pelos campos (nome + conteúdo), então reenviar o mesmo arquivo com outro boundary conta como igual.

### 2.2 Tabela de status HTTP

| Status | Quando |
|---|---|
| `200` | Leitura ou confirmação com sucesso |
| `201` | Pessoa ou transação criada (inclusive transação `cancel_notfound`) |
| `400` | Validação (campo faltando ou inválido, imagem inválida, JSON inválido) |
| `401` | Token administrativo ausente ou inválido (`/admin/*`) |
| `403` | Token administrativo sem `persons:delete` |
| `204` | Exclusão feita (`DELETE /admin/persons/{id}`) |
| `404` | Pessoa ou transação inexistente |
| `409` | Conflito: e-mail ou pessoa já cadastrados, transação não pendente, chave em processamento |
| `413` | Body maior que 10 MiB |
| `422` | Código de confirmação errado, ou Idempotency-Key reutilizada com outro payload |
| `500` | Erro interno (`{"error":"erro interno"}`) |
| `502` | Falha ao falar com o serviço de reconhecimento facial |

---

## 3. Endpoints

### 3.1 `POST /persons`: cadastrar pessoa

Cadastra a pessoa no serviço de reconhecimento facial parceiro, salva o identificador retornado
por ele, guarda a foto do nosso lado e armazena o código como hash (bcrypt).

**Headers**

| Header | Obrigatório | Valor |
|---|---|---|
| `Idempotency-Key` | sim | UUID |
| `Content-Type` | sim | `multipart/form-data` (o curl define sozinho com `-F`) |

**Body (multipart/form-data)**

| Campo | Tipo | Obrigatório | Regras |
|---|---|---|---|
| `name` | string | sim | Não vazio após trim, máx. 120 caracteres |
| `email` | string | sim | E-mail válido; salvo em minúsculas; único |
| `code` | string | não | 4 dígitos; padrão `"0000"` |
| `image` | arquivo | sim | JPEG/PNG até 8 MiB |

**curl**

```bash
curl -X POST http://localhost:8080/persons \
  -H "Idempotency-Key: 6f1c2a8e-3b7d-4c55-9a10-0e2f4b8d1a21" \
  -F name="Ana Souza" \
  -F email=ana@example.com \
  -F code=1234 \
  -F image=@foto.jpg
```

**201 Created**

```json
{
  "id": 1,
  "name": "Ana Souza",
  "email": "a***@example.com",
  "created_at": "2026-10-17T19:37:57.639Z"
}
```

**Erros**

| Status | Body | Causa |
|---|---|---|
| 400 | `{"error":"name é obrigatório (máx. 120 caracteres)"}` | name inválido |
| 400 | `{"error":"email inválido"}` | email inválido |
| 400 | `{"error":"code deve ter exatamente 4 dígitos"}` | code inválido |
| 400 | `{"error":"campo image é obrigatório"}` | sem imagem |
| 400 | `{"error":"imagem deve ser JPEG ou PNG"}` | tipo inválido |
| 400 | `{"error":"imagem maior que 8 MiB"}` | imagem grande demais |
| 400 | `{"error":"imagem recusada pelo reconhecimento facial: Nenhum rosto detectado na imagem"}` | o parceiro não achou rosto na foto (a mensagem após `:` vem do parceiro) |
| 409 | `{"error":"email já cadastrado"}` | e-mail repetido |
| 409 | `{"error":"esta pessoa já está cadastrada"}` | o parceiro reconheceu um rosto que já existe na base |
| 502 | `{"error":"falha ao comunicar com o serviço de reconhecimento facial"}` | parceiro fora do ar ou com erro |

> A resposta **nunca** traz o código, o identificador do parceiro nem o caminho da foto; o e-mail vem mascarado.

---

### 3.2 `GET /persons/{id}`: consultar pessoa

Retorna somente dados públicos.

**Path params**

| Param | Tipo | Descrição |
|---|---|---|
| `id` | integer | ID da pessoa no Boopateway (campo `id` do cadastro) |

**curl**

```bash
curl http://localhost:8080/persons/1
```

**200 OK**

```json
{
  "id": 1,
  "name": "Ana Souza",
  "email": "a***@example.com",
  "created_at": "2026-10-17T19:37:57.639Z"
}
```

**Erros**

| Status | Body |
|---|---|
| 400 | `{"error":"id inválido"}` |
| 404 | `{"error":"pessoa não encontrada"}` |

---

### 3.3 `POST /payments`: criar pagamento
Recebe **valor + foto**, identifica a pessoa pela foto (`POST /identify` no parceiro, que só consulta e nunca cadastra rostos desconhecidos) e cria a transação.
Recebe **valor + foto**, identifica a pessoa pela foto e cria a transação.

- Pessoa encontrada → `status: "pending_confirmation"` e a resposta traz `person_name`.
- Pessoa não encontrada → `status: "cancel_notfound"` (sem `person_name`). **Também retorna 201**, porque a transação é registrada.

A transação guarda internamente: id, valor, id da pessoa no parceiro, id da pessoa do nosso lado e status.

**Headers**

| Header | Obrigatório | Valor |
|---|---|---|
| `Idempotency-Key` | sim | UUID (evita cobrança duplicada) |
| `Content-Type` | sim | `multipart/form-data` |

**Body (multipart/form-data)**

| Campo | Tipo | Obrigatório | Regras |
|---|---|---|---|
| `amount_cents` | integer (enviado como texto) | sim | `> 0` e `<= 10000000000` |
| `image` | arquivo | sim | JPEG/PNG até 8 MiB |

**curl**

```bash
curl -X POST http://localhost:8080/payments \
  -H "Idempotency-Key: 0b9e7f3c-5d2a-4e61-8c47-2a1f9d6e3b08" \
  -F amount_cents=1990 \
  -F image=@foto.jpg
```

**201 Created (pessoa encontrada)**

```json
{
  "transaction_id": "txn_U54HQB2Z2LALAXGBHDAQQCBJKZ",
  "amount_cents": 1990,
  "status": "pending_confirmation",
  "person_name": "Ana Souza"
}
```

**201 Created (pessoa não encontrada)**

```json
{
  "transaction_id": "txn_X36J4Z4FKVGV6EYLARFZNKVFAG",
  "amount_cents": 1990,
  "status": "cancel_notfound"
}
```

**Erros**

| Status | Body | Causa |
|---|---|---|
| 400 | `{"error":"amount_cents deve ser um inteiro positivo (em centavos)"}` | valor ausente ou inválido |
| 400 | `{"error":"campo image é obrigatório"}` | sem imagem |
| 400 | `{"error":"imagem deve ser JPEG ou PNG"}` | tipo inválido |
| 400 | `{"error":"imagem recusada pelo reconhecimento facial: Nenhum rosto detectado na imagem"}` | o parceiro não achou rosto na foto (nenhuma transação é criada) |
| 502 | `{"error":"falha ao comunicar com o serviço de reconhecimento facial"}` | parceiro com erro (nenhuma transação é criada) |

> O cliente deve olhar o campo `status`, não só o HTTP 201, para saber se a pessoa foi encontrada.

---

### 3.4 `GET /payments/{id}`: consultar pagamento

**Path params**

| Param | Tipo | Descrição |
|---|---|---|
| `id` | string | `transaction_id` (prefixo `txn_`) |

**curl**

```bash
curl http://localhost:8080/payments/txn_U54HQB2Z2LALAXGBHDAQQCBJKZ
```

**200 OK**

```json
{
  "transaction_id": "txn_U54HQB2Z2LALAXGBHDAQQCBJKZ",
  "amount_cents": 1990,
  "status": "confirmed"
}
```

> Este endpoint não retorna `person_name`.

**Erros**

| Status | Body |
|---|---|
| 404 | `{"error":"transação não encontrada"}` |

---

### 3.5 `POST /payments/{id}/confirm`: confirmar pagamento

Confirma uma transação `pending_confirmation` com o código de 4 dígitos da pessoa.

**Headers**

| Header | Obrigatório | Valor |
|---|---|---|
| `Idempotency-Key` | sim | UUID **novo a cada tentativa de código** |
| `Content-Type` | sim | `application/json` |

**Path params**

| Param | Tipo | Descrição |
|---|---|---|
| `id` | string | `transaction_id` retornado em `POST /payments` |

**Body (JSON)**

```json
{ "code": "1234" }
```

| Campo | Tipo | Obrigatório | Regras |
|---|---|---|---|
| `code` | string | sim | Exatamente 4 dígitos. **Envie como string** (`"0012"`, não `12`) |

**curl**

```bash
curl -X POST http://localhost:8080/payments/txn_U54HQB2Z2LALAXGBHDAQQCBJKZ/confirm \
  -H "Idempotency-Key: 9a4d6b1e-7c3f-4f28-b5e0-6d8c2a1f7e94" \
  -H "Content-Type: application/json" \
  -d '{"code":"1234"}'
```

**200 OK**

```json
{
  "transaction_id": "txn_U54HQB2Z2LALAXGBHDAQQCBJKZ",
  "amount_cents": 1990,
  "status": "confirmed",
  "person_name": "Ana Souza"
}
```

**422 Unprocessable Entity (código errado)**

```json
{
  "error": "código inválido",
  "status": "pending_confirmation",
  "remaining_attempts": 2
}
```

Na 3ª tentativa errada:

```json
{
  "error": "código inválido",
  "status": "cancel_invalid_code",
  "remaining_attempts": 0
}
```

**Outros erros**

| Status | Body | Causa |
|---|---|---|
| 400 | `{"error":"envie {\"code\": \"<4 dígitos>\"}"}` | JSON inválido |
| 400 | `{"error":"code deve ter exatamente 4 dígitos"}` | code com formato errado (não conta como tentativa) |
| 404 | `{"error":"transação não encontrada"}` | id inexistente |
| 409 | `{"error":"transação não está pendente de confirmação"}` | já `confirmed`, `cancel_notfound`, `cancel_invalid_code` ou `cancel_person_deleted` |

---

### 3.6 Utilitários

```bash
curl http://localhost:8080/healthz
```

```json
{ "status": "ok" }
```

`GET /docs` serve a documentação interativa em HTML.

---

### 3.7 Endpoints administrativos (`/admin/*`)

Usados pelo portal administrativo (porta 31801). Todos exigem `Authorization: Bearer <token>`,
onde o token é o **token da API booprecognition**. A cada chamada o token é validado no
booprecognition; token recusado → `401 {"error":"token inválido"}`. Header ausente →
`401 {"error":"informe o token no header Authorization: Bearer <token>"}`.
Não exigem `Idempotency-Key`.

Pelo portal, as mesmas rotas ficam em `http://localhost:31801/api/admin/...` (proxy para a API).

#### `GET /admin/session`: validar token

```bash
curl http://localhost:8080/admin/session \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**200 OK**

```json
{ "scopes": ["identify", "persons:delete", "persons:write"], "can_delete": true }
```

`can_delete` é `true` quando o token tem o scope `persons:delete`.

#### `GET /admin/persons`: listar pessoas

Query: `page` (padrão 1), `limit` (padrão 20, máximo 100). Ordenado do mais recente para o mais antigo.

```bash
curl "http://localhost:8080/admin/persons?page=1&limit=20" \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**200 OK**

```json
{
  "items": [
    {
      "id": 1,
      "name": "Ana Souza",
      "email": "ana@example.com",
      "partner_person_id": 795820565,
      "created_at": "2026-10-17T21:52:35.804Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

O e-mail vem completo (área administrativa). O hash do código e o caminho da foto nunca são retornados.

#### `GET /admin/persons/{id}/image`: foto da pessoa

```bash
curl http://localhost:8080/admin/persons/1/image \
  -H "Authorization: Bearer $BOOP_TOKEN" -o foto.jpg
```

**200 OK** com o binário da imagem (`image/jpeg` ou `image/png`). `404` se a pessoa não existir.

#### `GET /admin/persons/{id}/transactions`: histórico de transações da pessoa

Query: `page` (padrão 1), `limit` (padrão 20, máximo 100). Ordenado da mais recente para a mais antiga.

```bash
curl "http://localhost:8080/admin/persons/1/transactions?page=1&limit=20" \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**200 OK**

```json
{
  "items": [
    {
      "transaction_id": "txn_AZVDZD5QAFUKDREAYS6L2NQZYZ",
      "amount_cents": 36448,
      "status": "confirmed",
      "attempts": 1,
      "created_at": "2026-10-17T22:03:16.591Z",
      "updated_at": "2026-10-17T22:03:23.108Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

`attempts` é o número de códigos errados digitados. `404 {"error":"pessoa não encontrada"}` se o id não existir.

#### `GET /admin/transactions`: buscar transações

Busca transações de todas as pessoas. Todos os filtros são opcionais e podem ser combinados (E lógico).

| Query | Tipo | Regra |
|---|---|---|
| `name` | string | Nome da pessoa **contém** o texto (sem diferenciar maiúsculas) |
| `email` | string | E-mail da pessoa **contém** o texto (sem diferenciar maiúsculas) |
| `amount_cents` | integer | Valor **exato** em centavos |
| `status` | string | `pending_confirmation`, `confirmed`, `cancel_notfound`, `cancel_invalid_code` ou `cancel_person_deleted` |
| `page` / `limit` | integer | Paginação (padrão 1 / 20, máx. 100) |

Filtrar por `name` ou `email` exclui transações sem pessoa vinculada (não identificadas ou de pessoa excluída).

```bash
curl "http://localhost:8080/admin/transactions?name=ana&status=confirmed&page=1&limit=20" \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**200 OK**

```json
{
  "items": [
    {
      "transaction_id": "txn_AZVDZD5QAFUKDREAYS6L2NQZYZ",
      "amount_cents": 36448,
      "status": "confirmed",
      "attempts": 1,
      "created_at": "2026-10-17T22:03:16.591Z",
      "updated_at": "2026-10-17T22:03:23.108Z",
      "person_id": 1,
      "person_name": "Ana Souza",
      "person_email": "ana@example.com",
      "partner_person_id": 795820565
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

`person_id`, `person_name` e `person_email` vêm `null` quando a transação não tem pessoa vinculada.

| Status | Body |
|---|---|
| 400 | `{"error":"status inválido"}` |
| 400 | `{"error":"amount_cents deve ser um inteiro positivo (em centavos)"}` |

#### `DELETE /admin/persons/{id}`: excluir pessoa

Exige o scope `persons:delete`. Remove a pessoa no booprecognition (`DELETE /persons/{partner_person_id}`),
na base e apaga a foto. **O histórico de transações é mantido**: as transações da pessoa perdem o
vínculo com ela (continuam com o ID no parceiro) e as `pending_confirmation` viram `cancel_person_deleted`.

```bash
curl -X DELETE http://localhost:8080/admin/persons/1 \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**204 No Content** (sem body)

| Status | Body | Causa |
|---|---|---|
| 400 | `{"error":"id inválido"}` | id não numérico |
| 401 | `{"error":"token inválido"}` | token recusado pelo booprecognition |
| 403 | `{"error":"token sem permissão para esta operação"}` | token sem `persons:delete` |
| 404 | `{"error":"pessoa não encontrada"}` | id inexistente |
| 502 | `{"error":"falha ao comunicar com o serviço de reconhecimento facial"}` | booprecognition fora do ar |

#### `GET /admin/logs`: logs de requisições

Toda chamada à API é registrada (exceto `/healthz`, `/docs` e o próprio `/admin/logs`) com método,
caminho, rota, query string, status, duração, IP, User-Agent e a mensagem de erro quando o status é
≥ 400. Corpos e headers **não** são gravados (contêm fotos, códigos e tokens). Os logs ficam 30 dias.
Mais recentes primeiro. Filtros opcionais: `method`, `path` (contém), `status` (`2xx`…`5xx`),
`from`/`to` (RFC 3339), `page`, `limit`.

```bash
curl "http://localhost:8080/admin/logs?status=4xx&path=/payments" \
  -H "Authorization: Bearer $BOOP_TOKEN"
```

**200 OK**
```json
{
  "items": [
    {
      "id": 2, "created_at": "2026-10-17T23:17:49.512Z", "method": "POST",
      "path": "/payments", "route": "/payments", "query": "", "status": 400, "duration_ms": 0,
      "client_ip": "172.18.0.1", "user_agent": "curl/8.19.0",
      "error": "header Idempotency-Key é obrigatório (máx. 255 caracteres)"
    }
  ],
  "total": 1, "page": 1, "limit": 20
}
```

---

## 4. Status da transação

| Status | Significado | Final? | Próxima ação |
|---|---|---|---|
| `pending_confirmation` | Pessoa identificada, aguardando código | não | `POST /payments/{id}/confirm` |
| `confirmed` | Pagamento confirmado | sim | — |
| `cancel_notfound` | Foto não reconhecida | sim | Criar novo pagamento (nova Idempotency-Key) |
| `cancel_invalid_code` | 3 códigos errados | sim | Criar novo pagamento |
| `cancel_person_deleted` | A pessoa foi excluída pelo portal enquanto o pagamento estava pendente | sim | — |

## 5. Roteiro completo (copiar e colar)

```bash
BASE=http://localhost:8080

# 1) Cadastrar pessoa
curl -X POST $BASE/persons \
  -H "Idempotency-Key: $(uuidgen)" \
  -F name="Ana Souza" \
  -F email=ana@example.com \
  -F code=1234 \
  -F image=@foto.jpg

# 2) Consultar pessoa
curl $BASE/persons/1

# 3) Criar pagamento (mesma pessoa da foto)
curl -X POST $BASE/payments \
  -H "Idempotency-Key: $(uuidgen)" \
  -F amount_cents=1990 \
  -F image=@foto.jpg
# -> copie o transaction_id da resposta

# 4) Confirmar com o código
curl -X POST $BASE/payments/txn_U54HQB2Z2LALAXGBHDAQQCBJKZ/confirm \
  -H "Idempotency-Key: $(uuidgen)" \
  -H "Content-Type: application/json" \
  -d '{"code":"1234"}'

# 5) Consultar status
curl $BASE/payments/txn_U54HQB2Z2LALAXGBHDAQQCBJKZ
```

Testando a idempotência (a 2ª chamada devolve a mesma resposta com `Idempotent-Replayed: true`
e não cria outra transação):

```bash
KEY=$(uuidgen)
curl -i -X POST $BASE/payments -H "Idempotency-Key: $KEY" -F amount_cents=1990 -F image=@foto.jpg
curl -i -X POST $BASE/payments -H "Idempotency-Key: $KEY" -F amount_cents=1990 -F image=@foto.jpg
```

## 6. Checklist para quem implementa um cliente

- [ ] Gerar `Idempotency-Key` (UUID) por operação e reutilizar só em retentativas da mesma operação.
- [ ] Repetir com a **mesma** chave em timeout, erro de rede e `5xx`; **não** repetir em `4xx`.
- [ ] Usar chave **nova** a cada nova tentativa de código no confirm.
- [ ] Enviar valores em centavos (inteiro).
- [ ] Enviar `code` como string de 4 dígitos.
- [ ] Após `POST /payments`, decidir pelo campo `status`, não só pelo HTTP 201.
- [ ] Exibir `remaining_attempts` quando o confirm retornar 422.
- [ ] Tratar `cancel_*` e `confirmed` como estados finais.
- [ ] Enviar a foto como JPEG/PNG de até 8 MiB no campo multipart `image`.
