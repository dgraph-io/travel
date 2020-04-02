package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"google.golang.org/grpc"
)

// Person ...
type Person struct {
	Name    string
	Balance int
}

// Query ...
func Query(ctx context.Context) error {
	gql := graphql.New("http://localhost:8080/query", http.DefaultClient)

	query := `{all(func: anyofterms(Name, "Bill")) {Name Balance}}`

	var resp struct {
		All []Person
	}
	err := gql.Query(context.TODO(), query, &resp)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
	return nil
}

// DB ...
func DB(ctx context.Context, host string) error {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)

	op := api.Operation{
		DropOp: api.Operation_ALL,
	}
	if err := client.Alter(ctx, &op); err != nil {
		return err
	}

	op = api.Operation{
		Schema: `
			Name: string @index(term) .
			Balance: int .
		`,
	}
	if err := client.Alter(context.Background(), &op); err != nil {
		return err
	}

	people := []Person{
		{Name: "Bill", Balance: 10},
		{Name: "Ale", Balance: 100},
	}
	out, err := json.Marshal(people)
	if err != nil {
		return err
	}

	txn := client.NewTxn()

	mut := api.Mutation{
		SetJson: out,
	}
	if _, err := txn.Mutate(ctx, &mut); err != nil {
		txn.Discard(ctx)
		return err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil
	}

	return nil
}
