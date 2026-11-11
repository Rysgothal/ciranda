# Ciranda

[Read in English (US)](README.en-US.md)

API para gerenciar o acervo, os empréstimos e as devoluções de uma biblioteca comunitária.

## Funcionalidades

- Cadastro, consulta, atualização e remoção de livros
- Registro de empréstimos e devoluções
- Controle de prazo e identificação de atrasos
- Listagem de livros com paginação e filtros
- Listagem de empréstimos em atraso
- Validações para impedir empréstimos de livros indisponíveis

## Stack

- Go
- SQLite
- [Framework Gin](github.com/gin-gonic/gin)
- Em desenvolvimento

## Organização esperada (em desenvolvimento)

A aplicação separa as responsabilidades entre handler, service e repository.

- **Handler:** recebe e responde às requisições HTTP
- **Service:** aplica as regras de empréstimo e devolução
- **Repository:** acessa os dados

As interfaces devem ser definidas pelo componente que as consome, sem adicionar camadas ou abstrações desnecessárias.

## Regras de negócio

- Um livro já emprestado não pode receber outro empréstimo até sua devolução
- Empréstimos e devoluções devem preservar a consistência do estado do livro
- Operações de empréstimo e devolução devem ser executadas em transação para evitar conflitos em acessos simultâneos

## Qualidade

As regras de negócio devem ser cobertas por testes de service usando repositórios falsos, sem dependência direta do banco de dados. O sistema deve manter o estado consistente quando houver tentativas simultâneas de emprestar o mesmo livro.
