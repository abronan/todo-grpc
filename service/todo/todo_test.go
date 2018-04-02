package todo

import (
	"context"
	"testing"

	api "github.com/abronan/todo-grpc/api/todo/v1"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TodoSuite struct {
	suite.Suite
	Todo *Service
}

func TestTodoTestSuite(t *testing.T) {
	db := pg.Connect(&pg.Options{
		User:     "abronan",
		Database: "todo",
		Addr:     "0.0.0.0:5432",
	})
	suite.Run(t, &TodoSuite{
		Todo: &Service{DB: db},
	})
}

func (s *TodoSuite) SetupTest() {
	s.Todo.DB.DropTable(&api.Todo{}, &orm.DropTableOptions{IfExists: true})
	s.Todo.DB.CreateTable(&api.Todo{}, nil)
}

func (s *TodoSuite) TearDownTest() {
	s.Todo.DB.DropTable(&api.Todo{}, &orm.DropTableOptions{IfExists: true})
}

func (s *TodoSuite) TestCreateTodo() {
	rcreate, err := s.Todo.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: &api.Todo{
				Title:       "item_1",
				Description: "item desc 1",
			},
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")
}

func (s *TodoSuite) TestCreateTodos() {
	rcreate, err := s.Todo.CreateTodos(
		context.Background(),
		&api.CreateTodosRequest{
			Items: []*api.Todo{
				&api.Todo{
					Title:       "item_1",
					Description: "item desc 1",
				},
				&api.Todo{
					Title:       "item_2",
					Description: "item desc 2",
				},
			},
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	for _, id := range rcreate.Ids {
		assert.NotEqual(s.T(), id, "")
	}
}

func (s *TodoSuite) TestGetTodo() {
	item := &api.Todo{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Todo.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	rget, err := s.Todo.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rget)
	assert.NotNil(s.T(), rget.Item)
	assert.Equal(s.T(), rget.Item, item)
}

func (s *TodoSuite) TestDeleteTodo() {
	item := &api.Todo{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Todo.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	rdel, err := s.Todo.DeleteTodo(
		context.Background(),
		&api.DeleteTodoRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rdel)

	// Getting the todo item should fail this time
	rget, err := s.Todo.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), rget)
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "Could not retrieve item from the database: pg: no rows in result set")
}

func (s *TodoSuite) TestUpdateTodo() {
	item := &api.Todo{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Todo.CreateTodo(
		context.Background(),
		&api.CreateTodoRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	newItem := &api.Todo{
		Id:          id,
		Title:       "item 1 update",
		Description: "updated desc",
		Completed:   true,
	}

	rupdate, err := s.Todo.UpdateTodo(
		context.Background(),
		&api.UpdateTodoRequest{
			Item: newItem,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rupdate)

	// Getting the todo item should return the updated version
	rget, err := s.Todo.GetTodo(
		context.Background(),
		&api.GetTodoRequest{
			Id: id,
		},
	)
	assert.NotNil(s.T(), rget)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), rget.Item.Id, newItem.Id)
	assert.Equal(s.T(), rget.Item.Title, newItem.Title)
	assert.Equal(s.T(), rget.Item.Description, newItem.Description)
	assert.Equal(s.T(), rget.Item.Completed, newItem.Completed)
}

func (s *TodoSuite) TestUpdateTodos() {
	items := []*api.Todo{
		{
			Title:       "item_1",
			Description: "item desc 1",
		},
		{
			Title:       "item_2",
			Description: "item desc 2",
		},
	}

	// Create the todo items
	resp, err := s.Todo.CreateTodos(
		context.Background(),
		&api.CreateTodosRequest{
			Items: items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), resp)

	// List the items and update their fields
	rlist, err := s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)

	for _, item := range rlist.Items {
		item.Description = "updated desc"
		item.Completed = true
	}

	rupdate, err := s.Todo.UpdateTodos(
		context.Background(),
		&api.UpdateTodosRequest{
			Items: rlist.Items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rupdate)

	// List again and see if the entries have had their fields changed
	rlist, err = s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)

	for _, item := range rlist.Items {
		assert.Equal(s.T(), item.Description, "updated desc")
		assert.True(s.T(), item.Completed)
	}
}

func (s *TodoSuite) TestListTodo() {
	items := []*api.Todo{
		{
			Title:       "item_1",
			Description: "item desc 1",
			Completed:   true,
		},
		{
			Title:       "item_2",
			Description: "item desc 2",
		},
		{
			Title:       "item_3",
			Description: "item desc 3",
		},
		{
			Title:       "item_4",
			Description: "item desc 4",
			Completed:   true,
		},
	}

	// List with empty database
	rlist, err := s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.Nil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 0)

	// Create the todo items
	rcreate, err := s.Todo.CreateTodos(
		context.Background(),
		&api.CreateTodosRequest{
			Items: items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)

	// List the items
	rlist, err = s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 4)

	// Limit the result of List
	rlist, err = s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{
			Limit: 2,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 2)

	// Only list non completed items
	rlist, err = s.Todo.ListTodo(
		context.Background(),
		&api.ListTodoRequest{
			NotCompleted: true,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 2)
}
