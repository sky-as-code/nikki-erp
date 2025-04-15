package core

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/sky-as-code/nikki-erp/modules/core/domain/user"
)

func SetupCQRS() (*cqrs.CommandBus, *cqrs.EventBus, error) {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	if err != nil {
		return nil, nil, err
	}

	router.AddMiddleware(middleware.CorrelationID)

	cqrsMarshaler := cqrs.JSONMarshaler{}

	commandBus, err := cqrs.NewCommandBus(
		pubSub,
		cqrs.CommandBusConfig{
			GenerateCommandID: watermill.NewUUID,
			Marshaler:         cqrsMarshaler,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	eventBus, err := cqrs.NewEventBus(
		pubSub,
		cqrs.EventBusConfig{
			GenerateEventID: watermill.NewUUID,
			Marshaler:       cqrsMarshaler,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return commandBus, eventBus, nil
}

func SetupUserModule(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, repo user.Repository) (*user.Service, error) {
	handler := user.NewCommandHandler(repo, eventBus)

	commandProcessor, err := cqrs.NewCommandProcessor(commandBus, cqrs.CommandProcessorConfig{})
	if err != nil {
		return nil, err
	}

	err = commandProcessor.AddHandlers(
		cqrs.NewCommandHandler(
			"CreateUser",
			handler.HandleCreateUser,
		),
		cqrs.NewCommandHandler(
			"UpdateUser",
			handler.HandleUpdateUser,
		),
		cqrs.NewCommandHandler(
			"DeleteUser",
			handler.HandleDeleteUser,
		),
		cqrs.NewCommandHandler(
			"GetUserByID",
			handler.HandleGetUserByID,
		),
		cqrs.NewCommandHandler(
			"GetUserByUsername",
			handler.HandleGetUserByUsername,
		),
		cqrs.NewCommandHandler(
			"GetUserByEmail",
			handler.HandleGetUserByEmail,
		),
	)
	if err != nil {
		return nil, err
	}

	return user.NewService(commandBus), nil
}
