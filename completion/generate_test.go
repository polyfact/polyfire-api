package completion

import (
	"context"
	"testing"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func MockLogRequests(
	_ string,
	_ string,
	_ string,
	_ string,
	_ int,
	_ int,
	_ database.Kind,
	_ bool,
) {
}

func TestSimpleFullGeneration(t *testing.T) {
	utils.SetLogLevel("WARN")
	ctx := context.Background()

	ctx = utils.MockOpenAIServer(ctx)

	ctx = context.WithValue(
		ctx,
		utils.ContextKeyDB,
		database.MockDatabase{
			/*
			 * This test is supposed to be the smallest generation possible. There shouldn't
			 * be a lot of database requests. If you need to add one here, be sure to check
			 * it's really useful, cannot be bypassed or merged with an already occuring request.
			 * Any database request will slow down requests for the users.
			 */
			MockLogRequests: MockLogRequests,
		},
	)

	/*
	 * This is normally set by the db/users.go request. Here we're only testing GenerationStart
	 * so we need to define it manually
	 */
	ctx = context.WithValue(ctx, utils.ContextKeyRateLimitStatus, database.RateLimitStatusOk)
	ctx = context.WithValue(ctx, utils.ContextKeyEventID, "00000000-0000-0000-0000-000000000000")

	reqBody := GenerateRequestBody{
		Task: "Test",
	}

	result, err := GenerationStart(ctx, "00000000-0000-0000-0000-000000000000", reqBody)
	if err != nil {
		t.Fatalf(`GenerationStart returned an error %v`, err)
	}

	str := ""

	for v := range *result {
		str += v.Result
	}

	if str != "Test response" {
		t.Fatalf(`Generate("Test") should have returned "Test response" but returned "%s"`, str)
	}
}
