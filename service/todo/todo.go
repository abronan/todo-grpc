package todo

import (
	"context"

	todo "github.com/abronan/todo-grpc/api/todo/v1"
	"github.com/go-pg/pg"
	"github.com/gogo/protobuf/types"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Service is the service dealing with storing
// and retrieving todo items from the database.
type Service struct {
	DB *pg.DB
}

// CreateTodo creates a todo given a description
func (s Service) CreateTodo(ctx context.Context, req *todo.CreateTodoRequest) (*todo.CreateTodoResponse, error) {
	req.Item.Id = uuid.NewV4().String()
	err := s.DB.Insert(req.Item)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert item into the database: %s", err)
	}
	return &todo.CreateTodoResponse{Id: req.Item.Id}, nil
}

// CreateTodos create todo items from a list of todo descriptions
func (s Service) CreateTodos(ctx context.Context, req *todo.CreateTodosRequest) (*todo.CreateTodosResponse, error) {
	var ids []string
	for _, item := range req.Items {
		item.Id = uuid.NewV4().String()
		ids = append(ids, item.Id)
	}
	err := s.DB.Insert(&req.Items)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert items into the database: %s", err)
	}
	return &todo.CreateTodosResponse{Ids: ids}, nil
}

// GetTodo retrieves a todo item from its ID
func (s Service) GetTodo(ctx context.Context, req *todo.GetTodoRequest) (*todo.GetTodoResponse, error) {
	var item todo.Todo
	err := s.DB.Model(&item).Where("id = ?", req.Id).First()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not retrieve item from the database: %s", err)
	}
	return &todo.GetTodoResponse{Item: &item}, nil
}

// ListTodo retrieves a todo item from its ID
func (s Service) ListTodo(ctx context.Context, req *todo.ListTodoRequest) (*todo.ListTodoResponse, error) {
	var items []*todo.Todo
	query := s.DB.Model(&items).Order("created_at ASC")
	if req.Limit > 0 {
		query.Limit(int(req.Limit))
	}
	if req.NotCompleted {
		query.Where("completed = false")
	}
	err := query.Select()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not list items from the database: %s", err)
	}
	return &todo.ListTodoResponse{Items: items}, nil
}

// DeleteTodo deletes a todo given an ID
func (s Service) DeleteTodo(ctx context.Context, req *todo.DeleteTodoRequest) (*todo.DeleteTodoResponse, error) {
	err := s.DB.Delete(&todo.Todo{Id: req.Id})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not delete item from the database: %s", err)
	}
	return &todo.DeleteTodoResponse{}, nil
}

// UpdateTodo updates a todo item
func (s Service) UpdateTodo(ctx context.Context, req *todo.UpdateTodoRequest) (*todo.UpdateTodoResponse, error) {
	req.Item.UpdatedAt = types.TimestampNow()
	res, err := s.DB.Model(req.Item).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update item: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update item from the database: %s", err)
	}
	return &todo.UpdateTodoResponse{}, nil
}

// UpdateTodos updates todo items given their respective title and description.
func (s Service) UpdateTodos(ctx context.Context, req *todo.UpdateTodosRequest) (*todo.UpdateTodosResponse, error) {
	time := types.TimestampNow()
	for _, item := range req.Items {
		item.UpdatedAt = time
	}
	res, err := s.DB.Model(&req.Items).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update items: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update items from the database: %s", err)
	}
	return &todo.UpdateTodosResponse{}, nil
}
