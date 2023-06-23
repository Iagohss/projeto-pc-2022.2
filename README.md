# Projeto de Programação Concorrente 2022.2
 
[Código base](https://github.com/pedrohenrique-ql/concorrente-lab-base).

## 🛠️ Construído com

- [Go 1.20.5](https://go.dev/learn/) - Linguagem de implementação
 
## ✒️ Autores

*  **Amilton Cristian** - *Desenvolvedor especializado em programação sequencial.* - [AmiltonCabral](https://github.com/AmiltonCabral)
*  **Iago Silva** - *Desenvolvedor especializado em programação concorrente.* - [Iagohss](https://github.com/Iagohss)
*  **Joab Cesar** - *Responsável por testes, processamento e desenvolvimento auxiliar.* - [Joabcmp](https://github.com/joabcmp)

### 🍷🗿 Estrategias adotadas

- A concorrência é introduzida por meio do uso de Go Routines. Para cada id lido no arquivo actors.txt é criada uma Go Routine executando a função handleActor. Onde, uma requisição é feita ao banco de dados para obter os dados de um ator específico. Em seguida, é criada uma nova goroutine para cada filme associado a esse ator. Desse modo, os processamentos e requisições relacionados a cada ator podem ser feitos concorrentemente, aumentando o desempenho do sistema.

- Para evitar problemas de concorrência durante o processamento do ranking, foram utilizados canais e sync.WaitGroup. Cada goroutine responsável por obter a avaliação média de um ator escreve seu resultado em um canal results. Essa abordagem permite que os atores sejam recebidos de forma segura e síncrona. A chamada wgAVGs.Wait() garante que a thread principal aguarde até que todos os atores tenham sido processados antes de fechar o canal results. Essa ação indica que nenhuma outra goroutine enviará dados para o canal. No contexto da função ranking, isso permite que a iteração sobre o canal seja finalizada corretamente quando todos os atores forem processados.

- Para evitar problemas de concorrência durante o cálculo da média de avaliações dos atores, foi usada a mesma abordagem. Cada goroutine responsável por obter a avaliação de um filme escreve seu resultado em um canal ratings. O sync.WaitGroup é utilizado para aguardar a finalização de todas as goroutines que obtêm as avaliações antes de calcular a média no método getActorAVGRating.

## ⚙️ Demonstração de possiveis resultados
*Concurrent.go*
![image](https://github.com/Iagohss/lab-pc-2022.2/assets/72311157/c53c8bba-0728-417d-8a46-9ae0d1120f94)

*sequencial.go*
![image](https://github.com/Iagohss/projeto-pc-2022.2/assets/72311157/028458b1-0f08-4ffa-9cc9-11e5cdbf6b17)
