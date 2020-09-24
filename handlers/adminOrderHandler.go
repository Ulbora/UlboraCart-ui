package handlers

import (
	"net/http"
	"strconv"
	"sync"

	six910api "github.com/Ulbora/Six910API-Go"
	sdbi "github.com/Ulbora/six910-database-interface"
	"github.com/gorilla/mux"
)

/*
 Six910 is a shopping cart and E-commerce system.
 Copyright (C) 2020 Ulbora Labs LLC. (www.ulboralabs.com)
 All rights reserved.
 Copyright (C) 2020 Ken Williamson
 All rights reserved.
 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.
 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.
 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

//OrderItem OrderItem
type OrderItem struct {
	ID                    int64   `json:"id"`
	Quantity              int64   `json:"quantity"`
	BackOrdered           bool    `json:"backOrdered"`
	Dropship              bool    `json:"dropship"`
	ProductName           string  `json:"productName"`
	ProductShortDesc      string  `json:"productShortDesc"`
	ProductID             int64   `json:"productId"`
	Sku                   string  `json:"sku"`
	OrderID               int64   `json:"orderId"`
	SpecialProcessing     bool    `json:"specialProcessing"`
	SpecialProcessingType string  `json:"specialProcessingType"`
	Price                 float64 `json:"price"`
	SalePrice             float64 `json:"salePrice"`
}

//OrderPage OrderPage
type OrderPage struct {
	Error            string
	Order            *sdbi.Order
	Notes            *[]sdbi.OrderComment
	OrderItemList    *[]OrderItem
	Orders           *[]sdbi.Order
	Status           string
	OrderStatusList  []string
	UserNameForNotes string
}

//StoreAdminEditOrderPage StoreAdminEditOrderPage
func (h *Six910Handler) StoreAdminEditOrderPage(w http.ResponseWriter, r *http.Request) {
	s, suc := h.getSession(r)
	h.Log.Debug("session suc in order edit view", suc)
	if suc {
		if h.isStoreAdminLoggedIn(s) {
			var oilist []OrderItem
			hd := h.getHeader(s)
			eovars := mux.Vars(r)
			idstr := eovars["id"]
			oID, _ := strconv.ParseInt(idstr, 10, 64)
			h.Log.Debug("order id in edit", oID)

			odErr := r.URL.Query().Get("error")
			var eoparm OrderPage
			eoparm.Error = odErr

			var wg sync.WaitGroup

			wg.Add(1)
			go func(oid int64, header *six910api.Headers) {
				defer wg.Done()
				odr := h.API.GetOrder(oid, header)
				h.Log.Debug("order in edit", odr)
				eoparm.Order = odr
			}(oID, hd)

			oItemList := h.API.GetOrderItemList(oID, hd)
			var oichan = make(chan OrderItem, len(*oItemList))
			for i := range *oItemList {
				wg.Add(1)
				go func(oi sdbi.OrderItem, ch chan OrderItem, header *six910api.Headers) {
					defer wg.Done()
					prod := h.API.GetProductByID(oi.ProductID, header)
					h.Log.Debug("prod in edit", prod)
					var noi OrderItem
					noi.ID = oi.ID
					noi.OrderID = oi.OrderID
					noi.BackOrdered = oi.BackOrdered
					noi.Dropship = oi.Dropship
					noi.ProductID = oi.ProductID
					noi.Sku = prod.Sku
					noi.ProductName = oi.ProductName
					noi.ProductShortDesc = oi.ProductShortDesc
					noi.Quantity = oi.Quantity
					noi.SpecialProcessing = prod.SpecialProcessing
					noi.SpecialProcessingType = prod.SpecialProcessingType
					noi.Price = prod.Price
					noi.SalePrice = prod.SalePrice
					ch <- noi

				}((*oItemList)[i], oichan, hd)
			}

			wg.Add(1)
			go func(oid int64, header *six910api.Headers) {
				defer wg.Done()
				notes := h.API.GetOrderCommentList(oID, hd)
				h.Log.Debug("notes in edit", notes)
				eoparm.Notes = notes
			}(oID, hd)

			wg.Wait()

			close(oichan)
			for coi := range oichan {
				oilist = append(oilist, coi)
			}

			eoparm.OrderItemList = &oilist
			eoparm.UserNameForNotes = usernameForAddedNotes

			eoparm.OrderStatusList = []string{"New", "Processing", "Not Paid", "Shipped", "Canceled", "Partial Cancel"}
			h.AdminTemplates.ExecuteTemplate(w, adminEditOrderPage, &eoparm)
		} else {
			http.Redirect(w, r, adminLogin, http.StatusFound)
		}
	}
}

//StoreAdminEditOrder StoreAdminEditOrder
func (h *Six910Handler) StoreAdminEditOrder(w http.ResponseWriter, r *http.Request) {
	s, suc := h.getSession(r)
	h.Log.Debug("session suc in prod edit", suc)
	if suc {
		if h.isStoreAdminLoggedIn(s) {
			eop := h.processOrder(r)
			found, eocom := h.processOrderComment(r)
			h.Log.Debug("order in update", *eop)
			hd := h.getHeader(s)
			res := h.API.UpdateOrder(eop, hd)
			if found {
				cres := h.API.AddOrderComments(eocom, hd)
				h.Log.Debug("order comment add resp", *cres)
			}
			h.Log.Debug("order update resp", *res)
			if res.Success {
				http.Redirect(w, r, adminOrderListView, http.StatusFound)
			} else {
				http.Redirect(w, r, adminOrderListViewFail, http.StatusFound)
			}
		} else {
			http.Redirect(w, r, adminLogin, http.StatusFound)
		}
	}
}

//StoreAdminViewOrderList StoreAdminViewOrderList
func (h *Six910Handler) StoreAdminViewOrderList(w http.ResponseWriter, r *http.Request) {
	s, suc := h.getSession(r)
	h.Log.Debug("session suc in prod list view", suc)
	if suc {
		if h.isStoreAdminLoggedIn(s) {
			hd := h.getHeader(s)
			volvars := mux.Vars(r)
			status := volvars["status"]
			var orders *[]sdbi.Order
			if status != "" {
				orders = h.API.GetStoreOrderListByStatus(status, hd)
			} else {
				orders = h.API.GetStoreOrderList(hd)
			}
			plErr := r.URL.Query().Get("error")
			var plparm OrderPage
			plparm.Error = plErr
			plparm.Orders = orders
			plparm.Status = status
			h.Log.Debug("orders  in list", orders)
			h.AdminTemplates.ExecuteTemplate(w, adminOrderListPage, &plparm)
		} else {
			http.Redirect(w, r, adminLogin, http.StatusFound)
		}
	}
}

func (h *Six910Handler) processOrder(r *http.Request) *sdbi.Order {
	var p sdbi.Order
	id := r.FormValue("id")
	p.ID, _ = strconv.ParseInt(id, 10, 64)
	p.Status = r.FormValue("status")
	subTotal := r.FormValue("subTotal")
	p.Subtotal, _ = strconv.ParseFloat(subTotal, 64)
	shippingHandling := r.FormValue("shippingHandling")
	p.ShippingHandling, _ = strconv.ParseFloat(shippingHandling, 64)
	insurance := r.FormValue("insurance")
	p.Insurance, _ = strconv.ParseFloat(insurance, 64)
	taxes := r.FormValue("taxes")
	p.Taxes, _ = strconv.ParseFloat(taxes, 64)
	total := r.FormValue("total")
	p.Total, _ = strconv.ParseFloat(total, 64)
	refunded := r.FormValue("refunded")
	p.Refunded, _ = strconv.ParseFloat(refunded, 64)
	p.OrderNumber = r.FormValue("orderNumber")
	p.OrderType = r.FormValue("orderType")
	pickup := r.FormValue("pickup")
	p.Pickup, _ = strconv.ParseBool(pickup)
	p.Username = r.FormValue("username")
	p.CustomerName = r.FormValue("customerName")
	customerID := r.FormValue("customerId")
	p.CustomerID, _ = strconv.ParseInt(customerID, 10, 64)
	p.BillingAddress = r.FormValue("billingAddress")
	billingAddressID := r.FormValue("billingAddressId")
	p.BillingAddressID, _ = strconv.ParseInt(billingAddressID, 10, 64)
	p.ShippingAddress = r.FormValue("shippingAddress")
	shippingAddressID := r.FormValue("shippingAddressId")
	p.ShippingAddressID, _ = strconv.ParseInt(shippingAddressID, 10, 64)
	shippingMethodID := r.FormValue("shippingMethodId")
	p.ShippingMethodID, _ = strconv.ParseInt(shippingMethodID, 10, 64)
	p.ShippingMethodName = r.FormValue("billingMethodName")
	storeID := r.FormValue("storeId")
	p.StoreID, _ = strconv.ParseInt(storeID, 10, 64)

	return &p
}

func (h *Six910Handler) processOrderComment(r *http.Request) (bool, *sdbi.OrderComment) {
	var c sdbi.OrderComment
	var found bool
	var com = r.FormValue("newComment")
	if com != "" {
		c.Comment = com
		c.Username = r.FormValue("usernameForNotes")
		oid := r.FormValue("id")
		c.OrderID, _ = strconv.ParseInt(oid, 10, 64)
		found = true
	}
	return found, &c
}
