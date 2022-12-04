package delivery

import (
	"github.com/spf13/cast"
	"net/http"
	"order-service/internal/data"
	"order-service/internal/service/helpers"
	requests "order-service/internal/service/requests/delivery"
	"order-service/resources"
	"strconv"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CreateDelivery(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCreateDeliveryRequest(r)
	if err != nil {
		helpers.Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	Delivery := data.Delivery{
		DeliveryPrice: request.Data.Attributes.DeliveryPrice,
		DeliveryDate:  request.Data.Attributes.DeliveryDate,
		OrderId:       cast.ToInt64(request.Data.Relationships.Order.Data.ID),
		AddressId:     cast.ToInt64(request.Data.Relationships.Address.Data.ID),
		StaffId:       cast.ToInt64(request.Data.Relationships.Staff.Data.ID),
	}

	var resultDelivery data.Delivery
	relateOrder, err := helpers.OrdersQ(r).FilterById(Delivery.OrderId).Get()
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to get order")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	relateAddress, err := helpers.AddressesQ(r).FilterById(Delivery.AddressId).Get()
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to get address")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	resultDelivery, err = helpers.DeliveriesQ(r).Insert(Delivery)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to create delivery")
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
			Customer: resources.Relation{
				Data: &resources.Key{
					ID:   strconv.FormatInt(relateOrder.CustomerId, 10),
					Type: resources.CUSTOMER_REF,
				},
			},
			Staff: resources.Relation{
				Data: &resources.Key{
					ID:   strconv.FormatInt(relateOrder.StaffId, 10),
					Type: resources.STAFF_REF,
				},
			},
			Cafe: resources.Relation{
				Data: &resources.Key{
					ID:   strconv.FormatInt(relateOrder.CafeId, 10),
					Type: resources.CAFE_REF,
				},
			},
		},
	})

	includes.Add(&resources.Address{
		Key: resources.NewKeyInt64(relateAddress.Id, resources.ADDRESS),
		Attributes: resources.AddressAttributes{
			BuildingNumber: relateAddress.BuildingNum,
			Street:         relateAddress.Street,
			City:           relateAddress.City,
			District:       relateAddress.District,
			Region:         relateAddress.Region,
			PostalCode:     relateAddress.PostalCode,
		},
	})

	result := resources.DeliveryResponse{
		Data: resources.Delivery{
			Key: resources.NewKeyInt64(resultDelivery.Id, resources.DELIVERY),
			Attributes: resources.DeliveryAttributes{
				DeliveryPrice: resultDelivery.DeliveryPrice,
				DeliveryDate:  resultDelivery.DeliveryDate,
			},
			Relationships: resources.DeliveryRelationships{
				Order: resources.Relation{
					Data: &resources.Key{
						ID:   strconv.FormatInt(resultDelivery.OrderId, 10),
						Type: resources.ORDER,
					},
				},
				Address: resources.Relation{
					Data: &resources.Key{
						ID:   strconv.FormatInt(resultDelivery.AddressId, 10),
						Type: resources.ADDRESS,
					},
				},
				Staff: resources.Relation{
					Data: &resources.Key{
						ID:   strconv.FormatInt(resultDelivery.StaffId, 10),
						Type: resources.STAFF_REF,
					},
				},
			},
		},
		Included: includes,
	}
	ape.Render(w, result)
}
