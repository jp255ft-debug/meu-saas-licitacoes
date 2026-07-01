# Padrões de Código NestJS - OrkaPay

## Anti-Alucinação em NestJS/TypeScript
- **NUNCA invente decorators do NestJS que não existem.** Consulte https://docs.nestjs.com/
- **NUNCA invente métodos de módulos npm.** Verifique a documentação oficial.
- **NUNCA use `any` em TypeScript.** Use tipos explícitos ou `unknown`.
- **NUNCA invente propriedades em objetos de resposta de APIs externas.** Consulte a documentação.
- **NUNCA use `require()` — use `import`/`export` do ES Modules.**

## Estrutura de Módulos

### Organização
```
src/modules/payment/
├── payment.module.ts          # @Module({...})
├── controllers/
│   ├── payment.controller.ts   # Rotas REST
│   └── webhook.controller.ts   # Webhooks
├── dto/
│   ├── create-payment.dto.ts   # DTOs de entrada
│   └── payment-response.dto.ts # DTOs de saída
├── services/
│   └── payment.service.ts      # Lógica de aplicação
├── repositories/
│   └── payment.repository.ts   # Acesso a dados
└── entities/
    └── payment.entity.ts       # TypeORM Entity
```

### Regras de Código
- **SEMPRE** use DTOs para validação de entrada (`class-validator`)
- **SEMPRE** use `@ApiTags()`, `@ApiOperation()` para documentação Swagger
- **NUNCA** exponha entidades diretamente nas respostas — use DTOs de saída
- **SEMPRE** use `@Injectable()` para serviços e `@Controller()` para controllers
- **NUNCA** coloque lógica de negócio em controllers — delegue para services
- **SEMPRE** use `@UsePipes(new ValidationPipe())` para validar DTOs

### TypeScript
- **NUNCA** use `any` — prefira `unknown` se o tipo for desconhecido
- **SEMPRE** configure `strict: true` no `tsconfig.json`
- **SEMPRE** use `interface` para objetos, `type` para uniões/alias
- **NUNCA** use `enum` — prefira `const` objects com `as const`
- **SEMPRE** use `async/await` em vez de `.then()/.catch()`

### Testes
- **Use** `@nestjs/testing` com `Test.createTestingModule()`
- **Mocke** dependências com `jest.fn()` ou `jest.spyOn()`
- **NUNCA** faça chamadas reais a banco de dados em testes unitários
