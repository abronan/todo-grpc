package todo

import (
	"context"
	"testing"

	api "github.com/abronan/todo-grpc/api/todo/v1"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
)

var todoService *Service

func init() {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Database: "todo",
		Addr:     "0.0.0.0:5432",
	})

	db.CreateTable(&api.Todo{}, nil)
	todoService = &Service{DB: db}
}

func TestCreateTodo(t *testing.T) {
	resp, err := todoService.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: &api.Todo{
				Title:       "todo_test",
				Description: "this is a todo item",
			},
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, resp.Id, "")
}

func TestCreateTodos(t *testing.T) {
	resp, err := todoService.CreateTodos(
		context.Background(),
		&api.CreateTodosRequest{
			Items: []*api.Todo{
				&api.Todo{
					Title:       "todo_test",
					Description: "this is a todo item",
				},
				&api.Todo{
					Title:       "another_test",
					Description: "this is another todo item",
				},
			},
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	for _, id := range resp.Ids {
		assert.NotEqual(t, id, "")
	}
}

func TestGetTodo(t *testing.T) {
	item := &api.Todo{
		Title:       "get_test",
		Description: "this is a todo item for testing Get",
	}

	resp, err := todoService.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, resp.Id, "")

	id := resp.Id

	getResp, err := todoService.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, getResp)
	assert.NotNil(t, getResp.Item)
	assert.Equal(t, getResp.Item, item)
}

func TestDeleteTodo(t *testing.T) {
	item := &api.Todo{
		Title:       "delete_test",
		Description: "this is a todo item for testing Delete",
	}

	resp, err := todoService.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, resp.Id, "")

	id := resp.Id

	delResp, err := todoService.DeleteTodo(
		context.Background(),
		&api.DeleteTodoRequest{
			Id: id,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, delResp)

	// Getting the todo item should fail this time
	getResp, err := todoService.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.Nil(t, getResp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Could not retrieve item from the database: pg: no rows in result set")
}

func TestUpdateTodo(t *testing.T) {
	item := &api.Todo{
		Title:       "update_test",
		Description: "this is a todo item for testing Update",
	}

	resp, err := todoService.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, resp.Id, "")

	id := resp.Id

	upItem := &api.Todo{
		Id:          id,
		Title:       "updated_todo_item",
		Description: "this is an updated todo",
		Completed:   true,
	}

	upResp, err := todoService.UpdateTodo(
		context.Background(),
		&api.UpdateTodoRequest{
			Item: upItem,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, upResp)

	// Getting the todo item should return the updated version
	getResp, err := todoService.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.NotNil(t, getResp)
	assert.Nil(t, err)
	assert.Equal(t, getResp.Item.Id, upItem.Id)
	assert.Equal(t, getResp.Item.Title, upItem.Title)
	assert.Equal(t, getResp.Item.Description, upItem.Description)
	assert.Equal(t, getResp.Item.Completed, upItem.Completed)
}
