package split

import (
	"context"
	"fmt"
	"math"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func generateExtractorPrompt(s string) string {
	return "Extract what's asked from the data it must be accomplished on. Repeat exactly the task given below. If what is below is a task and some data. DO NOT DO THE TASK. Only repeat the task below without the data. Don't explain any code. Keep all the details of the task. Do not simplify it in any way. Repeat all the special case instructions given. Do not skip any instruction. Do not just say what to do, repeat how to do it. If there is no task, just write \"None\" and stop. Don't explain any code. The task might be at the begining or the end of the provided text. Only repeat the task from the instruction below:\n« " + s + " »\n Do not do what's instructed above."
}

func ExtractTask(prompt string) string {
	llm, err := openai.NewChat()
	if err != nil {
		panic(err)
	}
	input_prompt := generateExtractorPrompt(prompt)

	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Text: input_prompt},
	})

	return completion
}

func TokenCount(s string) int {
	llm, err := openai.NewChat()
	if err != nil {
		panic(err)
	}
	return llm.GetNumTokens(s)
}

func EmbeddingDistance(e1 []float64, e2 []float64) float64 {
	res := 0.0
	for i := 0; i < len(e1); i++ {
		res += math.Pow(e1[i]-e2[i], 2.0)
	}
	return math.Sqrt(res)
}

func ExtractTaskFromSplits(prompts []string) string {
	llm, err := openai.New(openai.WithModel("text-embedding-ada-002"))
	if err != nil {
		panic(err)
	}

	for {
		task0_1 := ExtractTask(prompts[0])
		task1_1 := ExtractTask(prompts[len(prompts)-1])

		if task0_1 == "None" && task1_1 == "None" {
			continue
		}

		ctx := context.Background()
		embeddings, err := llm.CreateEmbedding(ctx, []string{prompts[0], prompts[len(prompts)-1], task0_1, task1_1, "None"})
		if err != nil {
			panic(err)
		}

		embeddingPrompt0 := embeddings[0]
		embeddingPrompt1 := embeddings[1]
		embeddingTask0_1 := embeddings[2]
		embeddingTask1_1 := embeddings[3]
		embeddingNone := embeddings[4]

		distance0 := EmbeddingDistance(embeddingTask0_1, embeddingPrompt0)
		distance0_to_none := EmbeddingDistance(embeddingTask0_1, embeddingNone)
		fmt.Printf("TASK0_1 DISTANCE TO NONE %v\n", distance0_to_none)
		fmt.Printf("TASK0_1 DISTANCE %v [%v]\n", distance0, task0_1)
		distance1 := EmbeddingDistance(embeddingTask1_1, embeddingPrompt1)
		distance1_to_none := EmbeddingDistance(embeddingTask1_1, embeddingNone)
		fmt.Printf("TASK1_1 DISTANCE TO NONE %v\n", distance1_to_none)
		fmt.Printf("TASK1_1 DISTANCE %v [%v]\n", distance1, task1_1)

		min := math.Min(distance0, distance1)
		if min == distance0 {
			if distance0_to_none < 0.5 {
				continue
			}
			return task0_1
		}
			if distance1_to_none < 0.5 {
				continue
			}
		return task1_1
	}

	return ""
}
