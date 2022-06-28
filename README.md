## [fablo](https://github.com/hyperledger-labs/fablo)

###### Correr comandos na diretoria
> `/home/infos/stvgd/chaincodes/stvgd-chaincode`

**`> fablo`**
  - Ajuda binário.

**`> fablo recreate`**
- Desliga a rede.
- Remove os ficheiros gerados (`fablo-target`).
- Liga a rede com as configurações em `fablo-config.json` (caminho pré-definido é `$(pwd)/fablo-config.json`).

**`> fablo chaincode upgrade stvgd-chaincode <versão>`**
- Atualiza e instancia o chaincode nos peers da rede.
- Versão segue padrão `x.y.z` (última versão do chaincode atual no final do ficheiro `fablo-config.json`).

## [go](https://go.dev/doc/install)

> `bash: 'go' command not found`

Correr: `export PATH=$PATH:/usr/local/go/bin` e `go version` de seguida.

Caso o erro persista, proceder à reinstalação do `go`:
1. Na diretoria `~`, correr `sudo rm -rf /usr/local/go`.
2. De seguida, correr `sudo tar -C /usr/local -xzf go1.18.2.linux-amd64.tar.gz`.
3. E por final, tentar de novo `export PATH=$PATH:/usr/local/go/bin`.

## [Fablo REST](https://github.com/softwaremill/fablo-rest)

Importar o ficheiro `chaincodes/stvgd-chaincode/fablo-rest.postman_collection.json` para o Postman.

Selecionar o ambiente `fablo-rest` e verificar as seguintes variáveis de ambiente:
```txt
channel: my-channel1
chaincode: stvgd-chaincode
host:port: 20.224.242.57:8801
```

###### Pedidos HTTP

Para fazer um pedido HTTP à Fablo REST API], temos que executar o pedido **Enroll** 1º para gerar o token de autenticação para os restantes pedidos. O corpo do pedido deve conter:

```json
{
    "id": "admin",
    "secret": "adminpw"
}
```

Resposta esperada:
```json
✅ {"token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx-admin"}
```

Após obter o token, é possível fazer pedidos de métodos *CRUD* + *Queries* à API e os respetivos ativos
> **Batch & Activities** (***Production, Logistical***[*Reception, Registration, Transport*])

⚠️ O token expira de validade após alguns minutos e é necessário fazer o pedido **Enroll** ou **Reenroll** de novo

## Estrutura pedidos

##### ⚠️ Métodos CRUD das Logistical Activities ainda não estão finalizados! ⚠️

Todos os pedidos HTTP feitos à Fablo REST API são pedidos **POST**, exceto 1.

No corpo do pedido temos o JSON com a seguinte estrutura:

```json
{
    "method": "<ContractClass>:<Method>",
    "args": [<param1>, <param2>, <param3>, ...]
}
```

*ContractClass* é referente à struct `StvgdContract` em `stvgd-contract.go` e *Method* aos seus respetivos métodos. Em *args* temos um array de strings dos parâmetros do método da transação. Em parâmetros cujo argumento é um **objeto ou uma estrutura de dados mais complexa como um map, tem que se definir o JSON do objeto/estrutura stringificado/a (*stringify*).**
> É possível converter JSON para JSON stringified neste [link](https://onlinetexttools.com/json-stringify-text)

ℹ️ O campo *args* **é obrigatoriamente um array de strings.** Independentemente do tipo de dados dos atributos da função definida, no pedido são sempre enviadas strings. Os comentários nas templates JSON abaixo apenas indicam o tipo de dados dos campos na struct do modelo do chaincode.

---

### 📦 Batch

#### BatchExists ✔️

```json
{
    "method": "StvgdContract:BatchExists",
    "args": [
        batchID //string
    ]
}
```

#### CreateBatch ✔️

```json
{
    "method": "StvgdContract:CreateBatch",
    "args": [
        batchID,
        productionActivityID //(opt),
        productionUnitID,
        batchInternalID,
        supplierID,
        unit,
        batchTypeID //string,
        batchComposition //map[string]float32,
        quantity,
        ecs,
        ses //float32
    ]
}
```

#### ReadBatch ✔️

```json
{
    "method": "StvgdContract:ReadBatch",
    "args": [
        batchID //string
    ]
}
```

#### GetAllBatches ✔️

```json
{
    "method": "StvgdContract:GetAllBatches",
    "args": []
}
```

#### GetBatchHistory ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:GetBatchHistory",
    "args": [
        batchID //string
    ]
}
```

#### UpdateBatchQuantity ✔️

```json
{
    "method": "StvgdContract:UpdateBatchQuantity",
    "args": [
        batchID //string,
        newQuantity //float32
    ]
}
```

#### UpdateBatchInternalID ✔️

```json
{
    "method": "StvgdContract:UpdateBatchInternalID",
    "args": [
        batchID,
        newBatchInternalID //string
    ]
}
```

#### TransferBatch ✔️

```json
{
    "method": "StvgdContract:TransferBatch",
    "args": [
        batchID,
        newProductionUnitID //string
    ]
}
```

#### DeleteBatch ✔️

```json
{
    "method": "StvgdContract:DeleteBatch",
    "args": [
        batchID //string
    ]
}
```

#### DeleteAllBatches ✔️

```json
{
    "method": "StvgdContract:DeleteAllBatches",
    "args": []
}
```

---

### 🧵 Production Activity

#### ProductionActivityExists ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:ProductionActivityExists",
    "args": [
        productionActivityID //string
    ]
}
```

#### CreateProductionActivity ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:CreateProductionActivity",
    "args": [
        productionActivityID,
        productionUnitID,
        companyID,
        activityTypeID,
        activityStartDate,
        activityEndDate //string,
        inputBatches //map[string]float32,
        outputBatch,
        Batch,
        ECS,
        SES //float32
    ]
}
```

#### ReadProductionActivity ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:ReadProductionActivity",
    "args": [
        productionActivityID //string,
    ]
}
```

#### GetAllProductionActivities ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:GetAllProductionActivities",
    "args": []
}
```

#### DeleteProductionActivity ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:DeleteProductionActivity",
    "args": [
        productionActivityID //string,
    ]
}
```

#### DeleteAllProductionActivities ❔⚠️ (Por testar...)

```json
{
    "method": "StvgdContract:DeleteAllProductionActivities",
    "args": []
}
```

---

### 🚚 Logistical Activity Transport ❌

#### LogisticalActivityTransportExists ❌

```json
{
    ...
}
```

#### CreateLogisticalActivityTransport ❌

```json
{
    ...
}
```

#### ReadLogisticalActivityTransport ❌

```json
{
    ...
}
```

#### GetAllLogisticalActivitiesTransport ❌

```json
{
    ...
}
```

#### DeleteLogisticalActivityTransport ❌

```json
{
    ...
}
```

#### DeleteAllLogisticalActivitiesTransport ❌

```json
{
    ...
}
```

---

### 📋 Logistical Activity Registration ❌

#### LogisticalActivityRegistrationExists ❌

```json
{
    ...
}
```

#### CreateLogisticalActivityRegistration ❌

```json
{
    ...
}
```

#### ReadLogisticalActivityRegistration ❌

```json
{
    ...
}
```

#### GetAllLogisticalActivitiesRegistration ❌

```json
{
    ...
}
```

#### DeleteLogisticalActivityRegistration ❌

```json
{
    ...
}
```

#### DeleteAllLogisticalActivitiesRegistration ❌

```json
{
    ...
}
```

---

### 📥 Logistical Activity Reception ❌

#### LogisticalActivityReceptionExists ❌

```json
{
    ...
}
```

#### CreateLogisticalActivityReception ❌

```json
{
    ...
}
```

#### ReadLogisticalActivityReception ❌

```json
{
    ...
}
```

#### GetAllLogisticalActivitiesReception ❌

```json
{
    ...
}
```

#### DeleteLogisticalActivityReception ❌

```json
{
    ...
}
```

#### DeleteAllLogisticalActivitiesReception ❌

```json
{
    ...
}
```