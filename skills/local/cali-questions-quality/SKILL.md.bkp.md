---
name: cali-questions-quality
description: >
  Use esta skill sempre que o usuário quiser avaliar, revisar ou melhorar perguntas de entrevistas,
  roteiros terapêuticos, questionários qualitativos, pesquisas de narrativa, roteiros JTBD (Jobs to Be Done)
  ou qualquer conjunto de perguntas destinado a captar histórias reais e significados pessoais de pessoas.
  Acione também quando o usuário colar perguntas de artigos, arquivos, roteiros de pesquisa ou
  formulários e pedir para verificar se seguem boas práticas narrativas, ou quando pedir sugestões
  de perguntas que gerem histórias autênticas em vez de análises abstratas.
  Trigger words: "avaliar perguntas", "roteiro de entrevista", "captar narrativa", "pesquisa qualitativa",
  "perguntas para história", "revisar questionário", "perguntas terapêuticas", "narrative interview",
  "JTBD", "jobs to be done", "entrevista de cliente", "roteiro de pesquisa".
---

# Narrative Capture — Avaliação e Melhoria de Perguntas

Skill para avaliar se perguntas de entrevistas, roteiros e questionários seguem os princípios
de captação de narrativas e significação. Para cada pergunta analisada, a skill:

1. Detecta **ambiguidades e abstrações** (alerta imediato — prioridade máxima)
2. Verifica violação dos **14 princípios** (10 narrativos + 4 de Dave Snowden/Cynefin)
3. Propõe **alternativas otimizadas** que respeitam todos os princípios
4. **Re-avalia cada alternativa** antes de publicar (Princípio 10)

---

## 🚨 ALERTA DE AMBIGUIDADE — Prioridade Máxima

**Antes de avaliar qualquer outro princípio, verifique se a pergunta é ambígua ou abstrata.**

Uma pergunta é ambígua ou abstrata quando:
- Pode ser interpretada de formas muito diferentes por pessoas diferentes
- Usa palavras relativas sem referencial claro: "rápido", "fácil", "bom", "melhor", "confortável", "prático", "barato", "conveniente", "saudável"
- Usa expressões vagas de frequência ou intensidade: "às vezes", "geralmente", "costuma", "sempre", "nunca", "muito", "pouco"
- O entrevistado precisaria saber o que o entrevistador "quis dizer" para responder corretamente
- Conceitos sem definição operacional: "sucesso", "felicidade", "qualidade", "valor", "progresso", "melhora"

**Regra de ouro (Bob Moesta):**
> "There's no fast, just faster than…"
> "There's no easy, just easier than…"
> "There's no healthy, just healthier than…"

Palavras relativas não têm significado absoluto. Sem desembrulhar o referencial, a resposta é ruído —
cada pessoa interpreta de um jeito, tornando as respostas incomparáveis entre si.

**Como sinalizar no output:**
```
🚨 AMBIGUIDADE: a palavra "[termo vago]" não tem referencial concreto.
Risco: cada pessoa interpreta de um jeito diferente.
```

**Como corrigir ambiguidades — técnicas de "desembrulhamento":**
- Pedir o oposto: "O que seria o contrário disso pra você?"
- Pedir números/concretude: "Quanto tempo? Quantos passos? Quanto custou exatamente?"
- Comparar com o passado: "Comparado com o que você fazia antes, o que mudou?"
- Bracketing (extremos): "O que seria rápido demais? E o que seria lento demais?"
- Perguntar o que não é: "Me conta o que não seria [palavra vaga] nessa situação."
- Perguntar o referencial: "Rápido comparado com o quê? Quanto você esperava que demorasse?"

---

## Os 9 Princípios de Captação Narrativa

Cada pergunta pode violar um ou mais princípios. Avalie todos antes de classificar.

---

### Princípio 1 — Baixo Custo Cognitivo
**A pergunta deve ser intuitiva, respondível sem esforço filosófico ou analítico.**

❌ Viola quando:
- Exige raciocínio abstrato ou generalizações ("Como você lida com X?", "O que significa Y para você?")
- Pede reflexão filosófica sobre si mesmo ("Qual é a sua relação com Z?")
- Tem tom de interrogatório ou teste intelectual
- O entrevistado precisa "pensar bastante" antes de responder

✅ Passa quando:
- Qualquer pessoa consegue responder espontaneamente, sem pausa para "pensar na resposta certa"
- A pergunta ancora em algo concreto: um momento, uma pessoa, um lugar, uma ação

**Nota sobre exemplos:**
Exemplos servem para BAIXAR o custo cognitivo quando a pessoa está com tela em
branco — NÃO para oferecer opções fechadas.

❌ Exemplos como opções fechadas: "foi dormir? falou com alguém? ficou no celular?"
  → Risco de priming: a pessoa escolhe um exemplo em vez de descrever o que fez.

✅ Exemplos como convite para ir além: "...pode ser qualquer coisa — desde ficar
parada até tomar uma atitude."
  → Comunica: "não estou te dando opções, estou te convidando a descrever."

✅ Pergunta ancorada na sequência (alternativa a exemplos): "o que aconteceu logo
depois? me conta o que veio na sequência."
  → "O que aconteceu" pede narrativa, não análise. "Na sequência" é concreto.

Regra: exemplos devem ser verbos genéricos abertos ("descansar, falar, confiar"),
nunca cenários específicos fechados ("foi dormir, falou com alguém").

---

### Princípio 2 — Linguagem Intuitiva para o Público-Alvo
**Os termos usados devem ser naturais para quem responde, não para quem pesquisa.**

❌ Viola quando:
- Usa jargão técnico, clínico, acadêmico ou de autoajuda ("padrões de apego", "crença limitante", "ressignificação", "job to be done", "ancoragem", "push/pull", "gatilho emocional")
- Usa conceitos que só fazem sentido dentro de uma teoria específica
- O vocabulário pertence ao entrevistador, não ao mundo do entrevistado

✅ Passa quando:
- As palavras são as que o próprio entrevistado usaria para descrever sua vida
- Termos especializados são substituídos por descrições funcionais simples

---

### Princípio 3 — Sem Nomeação ou Classificação Abertas
**Não peça para a pessoa nomear, categorizar ou rotular algo diretamente.**

❌ Viola quando:
- Pede para dar nome a algo: "Como você chamaria esse padrão?", "O que você diria que é isso?"
- Pede para classificar: "Isso é um problema emocional ou comportamental?"
- Abre a pergunta com "O que é / Quem é você?"

✅ Passa quando:
- Antes de qualquer nomeação, perguntas sobre ações e efeitos pavimentam o caminho
- O nome emerge naturalmente das histórias contadas, não é exigido de saída
- Sequência correta: **AÇÃO → EFEITO → (opcional) NOME**

---

### Princípio 4 — Histórias Reais, Não Cenários Fabricados
**A pergunta deve pedir um evento específico que aconteceu, não uma descrição genérica de como as coisas costumam ser.**

❌ Viola quando:
- Começa com "Como costuma ser…", "Como você geralmente…", "Quando isso acontece, você…"
- Pede descrição de padrão: "Como o medo age na sua vida?"
- É ampla o suficiente para o respondente descrever "um jeito que as coisas são" em vez de "uma vez que aconteceu"

✅ Passa quando:
- Ancora no tempo: "a última vez", "a pior vez", "quando foi que…"
- Ancora em especificidade: dia, semana, situação concreta
- O entrevistado consegue narrar com começo, meio e fim

**Por que cenários são problemáticos:**
Cenários são resumos borrados de dezenas de memórias. Eles omitem detalhes, emoções e surpresas —
os elementos que tornam uma história útil. Uma história episódica tem enredo, pessoas específicas,
tempo e revelação de perspectiva.

**Dica de ouro (Bob Moesta):**
> "Conta pra mim sobre a última vez que [situação]…" — e siga o fluxo da história da pessoa.

#### Subprincípio 4.1 — Verifique demonstrativos antes de flaggar

Se a pergunta usa "essas situações", "aquele momento", "isso que você contou",
"naquela história" — o demonstrativo ANCORA no material já coletado. Verifique
se o referencial está claro no contexto da conversa antes de classificar como
cenário genérico.

Regra: se "essas situações" se refere a histórias já contadas pelo cliente na
mesma sessão, NÃO é cenário — é ancora.

---

### Princípio 5 — Evento Antes da Interpretação
**Perguntas de análise e interpretação só vêm depois que a história está sobre a mesa.**

❌ Viola quando:
- Pede análise antes de ter uma história: "O que você acha que causou esse padrão?"
- Mistura evento e interpretação numa mesma pergunta
- Convida à intelectualização antes do relato

✅ Sequência correta:
1. **Passo 1 — Pedir a história:** ancora no tempo/espaço, foca nos extremos ("a última vez", "a pior vez")
2. **Passo 2 — Pedir a interpretação:** depois que o episódio está narrado, convida à reflexão com distância narrativa

---

### Princípio 6 — Contraste Cria Significado
**Toda pergunta que busca entender motivação, valor ou preferência precisa de um ponto de comparação.**

> "Contrast creates meaning. Context creates value." — Bob Moesta

❌ Viola quando:
- Pede uma avaliação ou preferência sem referencial: "Por que você escolheu isso?"
- Assume que a pessoa sabe comparar sem ser provocada a isso
- A resposta caberia numa palavra sem revelar o contexto

✅ Passa quando:
- Coloca explicitamente uma alternativa ou oposto: "Por que você fez isso e não [alternativa]?"
- Força a pessoa a explicar o que ficou de fora da escolha
- Usa o passado como espelho do presente: "Comparado com como era antes, o que mudou?"

**Técnica do Bracketing:**
Ofereça dois extremos para a pessoa explicar o meio:
> "O que seria [extremo A]? E o que seria [extremo B]? Onde você acha que estava o ideal?"

---

### Princípio 7 — Âncora Temporal ("Por que agora?")
**Perguntas sobre decisão ou mudança devem sempre explorar o gatilho temporal.**

O "por que agora?" é um dos contrastes mais poderosos — revela o que mudou na vida da pessoa
que a fez se mover. Sem essa âncora, a resposta flutua sem contexto.

❌ Viola quando:
- Pergunta sobre decisões sem explorar o momento em que tudo mudou
- Ignora o que acontecia na vida da pessoa antes da mudança
- Não explora o que estava segurando a pessoa

✅ Exemplos de perguntas de âncora temporal:
- "Quando você começou a pensar nisso pela primeira vez?"
- "O que te impediu de fazer isso antes?"
- "O que mudou na sua vida que fez você decidir agora e não há seis meses?"
- "Por que você fez isso agora e não há um ano?"
- "O que estava acontecendo na sua vida naquele momento?"

---

### Princípio 8 — Explorar o Que Foi Rejeitado
**Entender o que a pessoa descartou revela tanto quanto entender o que ela escolheu.**

Perguntas sobre o que ficou de fora da decisão revelam o que realmente importou, o que foi
rejeitado e por quê — e os freios que quase impediram a mudança.

❌ Viola quando:
- Foca só no que foi escolhido, ignorando as alternativas consideradas
- Não explora a resistência ou os medos que quase impediram a decisão

✅ Exemplos de perguntas sobre o que foi rejeitado:
- "Por que você não continuou com o que tinha antes?"
- "O que te fez descartar [alternativa X]?"
- "Teve algum momento em que você quase desistiu? O que aconteceu?"
- "O que te deixava com medo ou receio de mudar?"
- "Por que não simplesmente [opção óbvia que a pessoa não escolheu]?"

**As 4 forças que moldam toda decisão de mudança:**
Um roteiro completo deve cobrir as quatro:
- **Push** — o que empurrou para fora da situação antiga (frustrações, dores acumuladas)
- **Pull** — o que atraiu para o novo (benefícios esperados, imagem desejada)
- **Ansiedade** — o que travou (medos, riscos percebidos, dúvidas sobre o novo)
- **Hábito** — o que segurou (lealdade ao status quo, conforto com o antigo)

---

### Princípio 9 — Concretude e Especificidade
**Perguntas devem forçar respostas com detalhes concretos: quem, o quê, quando, quanto, onde.**

❌ Viola quando:
- Aceita respostas de uma palavra ou expressão vaga
- Não convida a pessoa a descrever passo a passo o que aconteceu
- Perguntas abertas demais permitem respostas genéricas

✅ Exemplos de perguntas que induzem concretude:
- "Me conta passo a passo o que aconteceu naquele dia."
- "Quem mais estava envolvido? O que cada um disse?"
- "Quanto tempo você demorou para decidir? O que foi que aconteceu nesse tempo?"
- "O que você estava fazendo quando [situação] aconteceu?"

**Técnica: pedir o que NÃO É**
É mais fácil para as pessoas descrever pelo negativo. Quando uma resposta for vaga, pergunte:
> "Me conta o que não seria [palavra vaga] pra você nessa situação."

---

### Princípio 10 — Verificação de Alternativa
**Após propor uma alternativa, verifique se ela realmente corrige a violação original.**

**Checklist de verificação:**
- A alternativa usa o mesmo termo vago que foi flagged? Se sim, não resolveu.
- A alternativa introduziu novos termos vagos? Se sim, criou novo problema.
- A alternativa manteve a intenção original da pergunta? Se não, perdeu o dado que se buscava.
- Re-avalie a alternativa contra os mesmos 14 princípios. Se falha, reescreva de novo.

**Regra:** antes de publicar a alternativa, re-avalie-a. Uma alternativa que não
corrige o problema é pior que nenhuma alternativa — dá falsa sensação de correção.

---

## Princípios Complementares — Dave Snowden (Cynefin)

Estes princípios complementam os 10 princípios narrativos, especialmente em contextos de facilitação de grupos, estratégia e mudança organizacional. Use-os para avaliar perguntas de diálogo, sondas e exercícios coletivos.

---

### Princípio 11 — Possível Adjacente (Atravesse o rio tateando as pedras)
**Abandone projetos de estados futuros idealizados. Em complexidade, o futuro é imprevisível.**

❌ Viola quando:
- Pede visão de longo prazo abstrata: "Qual é sua visão para daqui a 5 anos?"
- Usa perguntas que forçam a pessoa a inventar uma utopia inatingível
- Tenta projetar o futuro paralisa a ação ou força os dados a caberem na teoria

✅ Passa quando:
- Foca no próximo passo prático a partir dos recursos reais de hoje
- Mantém o escopo da ação no presente comum

**Positivo:** "A partir de onde estamos hoje e dos recursos que temos agora, qual é o próximo passo prático que podemos dar amanhã de manhã?"

---

### Princípio 12 — Amplificação e Amortecimento (Mais destas, menos daquelas)
**Use as narrativas do dia a dia da própria comunidade como material bruto da estratégia.**

❌ Viola quando:
- Usa jargão corporativo vago: "cultura de excelência", "melhores práticas", "transformação digital"
- Pede para criar políticas abstratas do zero sem ancorar em histórias reais
- Generaliza sem olhar para os padrões concretos do sistema

✅ Passa quando:
- Parte das histórias reais já contadas pelo grupo
- Convida a identificar e intervir nos padrões existentes

**Positivo:** "Olhando para essas histórias que vocês mesmos contaram, como podemos agir para ter mais histórias como estas (positivas) e menos histórias como aquelas (negativas)?"

---

### Princípio 13 — Experimentação (Use sondas seguras para falhar em paralelo)
**Em sistemas complexos, não perca tempo discutindo qual especialista está certo.**

❌ Viola quando:
- Propõe um único piloto "definitivo" para resolver problema crônico
- Assume que análise prévia garantirá o resultado
- Finge que o fracasso não é uma opção

✅ Passa quando:
- Propõe múltiplos experimentos pequenos, contraditórios e simultâneos
- Aceita que falhar é tolerável e informativo
- O foco está em aprender sobre o problema, não em provar uma tese

**Positivo:** "Quais são 3 ou 4 experimentos pequenos, rápidos e contraditórios que podemos testar em paralelo nesta semana, onde falhar seja tolerável e nos ensine algo sobre o problema?"

---

### Princípio 14 — Desintermediação e Aporia (O significado pertence a quem viveu)
**O facilitador não deve se posicionar como avaliador da verdade.**

❌ Viola quando:
- Posiciona o facilitador como intérprete权威: "Deixe-me analisar sua resposta e explicar o que isso significa"
- Força clareza prematura ou dá "a" resposta
- Centraliza a interpretação em vez de distribuí-la

✅ Passa quando:
- Retira-se do centro da interpretação
- Usa ferramentas (como Tríades) para criar aporia (confusão deliberada e paradoxo)
- Convida os próprios participantes a tropeçarem em suas contradições

**Positivo:** "Aqui estão os dados e as histórias. Como vocês mesmos agrupariam isso e o que vocês acham que esses padrões significam?"

---

## Fluxo de Avaliação

### 1. Identificar o contexto
- Quem é o público que responderá?
- Qual é o objetivo da entrevista/pesquisa?
- As perguntas vêm de um arquivo/artigo ou foram digitadas diretamente?

### 2. Varredura de ambiguidades (prioridade máxima)
Antes dos princípios: varredura completa de palavras vagas e termos sem referencial.
Emita alertas 🚨 para cada termo problemático.

### 3. Avaliar cada pergunta contra os 14 princípios

### 4. Pontuar e classificar
- ✅ **Passa** — cumpre todos os princípios relevantes
- ⚠️ **Ajuste** — pequeno problema, facilmente corrigível
- ❌ **Reescrever** — viola um ou mais princípios de forma significativa

### 5. Propor alternativas corrigidas
Para cada ⚠️ ou ❌: alternativa que corrija a violação mantendo a intenção original.

### 6. Re-avaliar cada alternativa (Princípio 10)
Após propor alternativas, re-avalie cada uma contra os mesmos princípios.
Se a alternativa não corrige o problema, reescreva.

---

## Formato de Saída

```
---
**Pergunta [N]:** "[texto original]"

🚨 AMBIGUIDADE: [termos vagos e risco, se houver]

**Avaliação:** ✅ Passa | ⚠️ Ajuste | ❌ Reescrever

**Princípios violados:** [liste, ou "Nenhum"]

**Por quê:** [explicação concisa]

**Alternativa(s):**
  A) "[versão corrigida]" — [o que essa versão prioriza]
  B) "[versão alternativa]" — [diferença de abordagem] (se houver)
---
```

**Resumo Geral ao final:**
- Total / ✅ / ⚠️ / ❌
- Padrão de problema mais frequente
- Termos vagos recorrentes que precisam de protocolo de desembrulhamento
- Cobertura das 4 forças (Push/Pull/Ansiedade/Hábito) no roteiro completo
- Sugestão estrutural de sequenciamento

---

## Exemplos de Referência

### Exemplo A — Ambiguidade com Palavra Relativa

**Original:** "O processo foi rápido para você?"

**🚨 AMBIGUIDADE:** "rápido" não tem referencial. Rápido comparado com quê?

**Avaliação:** ❌ Reescrever — Ambiguidade + Princípio 6 (sem contraste)

**Alternativas:**
A) "Quanto tempo você esperava que ia demorar? E quanto tempo demorou de verdade?"
B) "O que seria 'demorado demais' pra você nessa situação? Como foi comparado a isso?"

---

### Exemplo B — Cenário Fabricado

**Original:** "Como o medo costuma agir na sua vida?"

**Avaliação:** ❌ Reescrever — Princípio 1 (custo cognitivo alto) + Princípio 4 (cenário)

**Alternativas:**
A) "Você consegue se lembrar da última vez em que quis fazer algo mas algo te segurou? O que aconteceu, passo a passo?"
*(Após a história:)* "Olhando para o que você me contou — o que você acha que estava tentando te dizer naquele momento?"

---

### Exemplo C — Sem Âncora Temporal e Sem Contraste

**Original:** "Por que você decidiu mudar?"

**Avaliação:** ⚠️ Ajuste — Princípio 6 (sem contraste) + Princípio 7 (sem âncora temporal)

**Alternativas:**
A) "O que mudou na sua vida que fez você decidir agora e não há seis meses?"
B) "O que te segurou todo esse tempo? E o que foi diferente dessa vez que fez você agir?"

---

### Exemplo D — Só Cobre o Pull, Ignora as Outras 3 Forças

**Original:** "O que te atraiu nessa opção?"

**Avaliação:** ⚠️ Ajuste — incompleta: cobre apenas Pull, ignora Push, Ansiedade e Hábito

**Alternativas:**
A) "Por que você escolheu isso em vez de continuar com o que tinha antes?" (cobre Push + Pull)
B) "Teve algum momento em que você quase não fez a mudança? O que quase te segurou?" (cobre Ansiedade + Hábito)

---

### Exemplo E — Jargão Técnico

**Original:** "Onde você acredita que adquiriu suas crenças limitantes sobre dinheiro?"

**Avaliação:** ❌ Reescrever — Princípio 2 (jargão) + Princípio 1 (análise abstrata de origem)

**Alternativa:**
A) "Você se lembra de alguma conversa sobre dinheiro quando era criança — com seus pais, avós, alguém que marcou — que ficou na sua cabeça até hoje? O que foi dito naquele momento?"

---

### Exemplo F — Pergunta que Passa em Tudo ✅

**Original:** "Conta pra mim sobre a última vez que essa situação apareceu — o que foi que aconteceu, passo a passo, desde o começo?"

**Avaliação:** ✅ Passa. Baixo custo cognitivo, sem jargão, sem nomeação, âncora temporal clara ("a última vez"),
pede episódio concreto com sequência, não pede interpretação prematura.

---

## Notas para o Avaliador (Claude)

- **Ambiguidade é prioridade zero:** sinalize palavras vagas antes de qualquer outra avaliação.
- **Seja generoso com perguntas ambíguas de contexto:** se pode ser lida como passando nos princípios dado o contexto, classifique como ⚠️ em vez de ❌.
- **Respeite a intenção:** a alternativa deve buscar o mesmo dado que a pergunta original tentava acessar — só mudar a rota.
- **Avalie o roteiro como um todo:** verifique se cobre as 4 forças (Push, Pull, Ansiedade, Hábito) e se a sequência respeita evento antes de interpretação.
- **Perguntas de aquecimento são exceção:** perguntas introdutórias simples não precisam passar pelo crivo completo.
- **Retrospectividade:** sinalize se o roteiro parece projetado para quem ainda não viveu a experiência — perguntas narrativas funcionam melhor com quem acabou de viver a mudança.
- **Após a coleta:** lembre o usuário que debater 30–60 min por entrevista é parte do processo. Padrões surgem de 5–10 entrevistas bem conduzidas.