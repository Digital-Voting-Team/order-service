package order_item

import (
	"github.com/spf13/cast"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/data"
	"order-service/internal/service/helpers"
	requests "order-service/internal/service/requests/order_item"
	"order-service/resources"
	"strconv"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CreateOrderItem(endpointsConfig *config.EndpointsConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := requests.NewCreateOrderItemRequest(r)
		if err != nil {
			helpers.Log(r).WithError(err).Info("wrong request")
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}

		mealId := cast.ToInt64(request.Data.Relationships.Meal.Data.ID)
		resultMeal, err := helpers.ValidateMeal(r.Header.Get("Authorization"), endpointsConfig.Endpoints["menu-service"], mealId)
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to get meal from DB")
			ape.Render(w, problems.InternalError())
			return
		}
		if resultMeal == nil {
			ape.Render(w, problems.NotFound())
			return
		}

		OrderItem := data.OrderItem{
			Quantity: request.Data.Attributes.Quantity,
			MealId:   cast.ToInt64(resultMeal.Data.ID),
			OrderId:  cast.ToInt64(request.Data.Relationships.Order.Data.ID),
		}

		var resultOrderItem data.OrderItem
		relateOrder, err := helpers.OrdersQ(r).FilterById(OrderItem.OrderId).Get()
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to get order")
			ape.RenderErr(w, problems.NotFound())
			return
		}

		resultOrderItem, err = helpers.OrderItemsQ(r).Insert(OrderItem)
		if err != nil {
			helpers.Log(r).WithError(err).Error("failed to create order item")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		var includes resources.Included
		includes.Add(&resources.Order{
			Key: resources.NewKeyInt64(relateOrder.Id, resources.ORDER),
			Attributes: resources.OrderAttributes{
				TotalPrice:    relateOrder.TotalPrice,
				PaymentMethod: relateOrder.PaymentMethod,
				IsTakeAway:    relateOrder.IsTakeAway,
				OrderDate:     relateOrder.OrderDate,
			},
			Relationships: resources.OrderRelationships{
				Status: resources.Relation{
					Data: &resources.Key{
						ID:   strconv.FormatInt(relateOrder.StatusId, 10),
						Type: resources.STATUS,
					},
				},
			},
		})

		result := resources.OrderItemResponse{
			Data: resources.OrderItem{
				Key: resources.NewKeyInt64(resultOrderItem.Id, resources.ORDER_ITEM),
				Attributes: resources.OrderItemAttributes{
					Quantity: resultOrderItem.Quantity,
				},
				Relationships: resources.OrderItemRelationships{
					Meal: resources.Relation{
						Data: &resources.Key{
							ID:   strconv.FormatInt(resultOrderItem.MealId, 10),
							Type: resources.MEAL_REF,
						},
					},
					Order: resources.Relation{
						Data: &resources.Key{
							ID:   strconv.FormatInt(resultOrderItem.OrderId, 10),
							Type: resources.ORDER,
						},
					},
				},
			},
			Included: includes,
		}
		ape.Render(w, result)
	}
}
