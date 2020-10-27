package handlers

import (
	"net/http"
	"strconv"

	m "github.com/Ulbora/Six910-ui/managers"

	api "github.com/Ulbora/Six910API-Go"
	sdbi "github.com/Ulbora/six910-database-interface"
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

//CreateCustomerAccountPage CreateCustomerAccountPage
func (h *Six910Handler) CreateCustomerAccountPage(w http.ResponseWriter, r *http.Request) {
	ccuss, suc := h.getUserSession(r)
	h.Log.Debug("session suc", ccuss)
	if suc {
		hd := h.getHeader(ccuss)
		var caocp CustomerPage
		ml := h.MenuService.GetMenuList()
		h.getCartTotal(ccuss, ml, hd)
		caocp.MenuList = ml
		caocp.StateList = h.StateService.GetStateList("states")
		h.Templates.ExecuteTemplate(w, customerCreatePage, caocp)
	}
}

//CreateCustomerAccount CreateCustomerAccount
func (h *Six910Handler) CreateCustomerAccount(w http.ResponseWriter, r *http.Request) {
	crcuss, suc := h.getUserSession(r)
	h.Log.Debug("session suc", crcuss)
	if suc {
		hd := h.getHeader(crcuss)
		//appvars := mux.Vars(r)
		email := r.FormValue("email")
		fcus := h.API.GetCustomer(email, hd)
		if fcus != nil && fcus.ID != 0 {
			http.Redirect(w, r, createCustomerViewFail, http.StatusFound)
		} else {
			var ca m.CustomerAccount
			var cus sdbi.Customer
			var adlst []sdbi.Address
			var usr api.User

			firstName := r.FormValue("firstName")
			lastName := r.FormValue("lastName")
			company := r.FormValue("company")
			city := r.FormValue("city")
			state := r.FormValue("state")
			zip := r.FormValue("zip")
			phone := r.FormValue("phone")
			password := r.FormValue("password")

			billAddress := r.FormValue("billAddress")
			billCity := r.FormValue("billCity")
			billState := r.FormValue("billState")
			billZip := r.FormValue("billZip")
			billCountry := r.FormValue("billCountry")
			if billAddress != "" && billCity != "" && billState != "" && billZip != "" {
				var ba sdbi.Address
				ba.Address = billAddress
				ba.City = billCity
				ba.State = billState
				ba.Zip = billZip
				ba.Country = billCountry
				ba.Type = billingAddressType
				adlst = append(adlst, ba)
			}

			shipAddress := r.FormValue("shipAddress")
			shipCity := r.FormValue("shipCity")
			shipState := r.FormValue("shipState")
			shipZip := r.FormValue("shipZip")
			shipCountry := r.FormValue("shipCountry")
			if shipAddress != "" && shipCity != "" && shipState != "" && shipZip != "" {
				var sa sdbi.Address
				sa.Address = shipAddress
				sa.City = shipCity
				sa.State = shipState
				sa.Zip = shipZip
				sa.Country = shipCountry
				sa.Type = shippingAddressType
				adlst = append(adlst, sa)
			}

			cus.Email = email
			cus.FirstName = firstName
			cus.LastName = lastName
			cus.Company = company
			cus.City = city
			cus.State = state
			cus.Zip = zip
			cus.Phone = phone

			usr.Username = email
			usr.Password = password
			usr.Role = customerRole
			usr.Enabled = true

			ca.Customer = &cus
			ca.Addresses = &adlst
			ca.User = &usr

			suc, cres := h.Manager.CreateCustomerAccount(&ca, hd)
			h.Log.Debug("cres suc: ", suc)
			h.Log.Debug("acres: ", cres)
			//acres := h.storeCustomerCart(cres, cpls, w, r)
			if suc {
				h.Log.Debug("acres customer: ", cres.Customer)
				h.Log.Debug("acres customer id: ", cres.Customer.ID)

				crcuss.Values["username"] = email
				crcuss.Values["password"] = password
				crcuss.Values["loggedIn"] = true
				crcuss.Values["customerUser"] = true
				serr := crcuss.Save(r, w)
				h.Log.Debug("serr", serr)
				cc := h.getCustomerCart(crcuss)
				cc.CustomerAccount = cres
				h.storeCustomerCart(cc, crcuss, w, r)

				http.Redirect(w, r, customerShoppingCartView, http.StatusFound)
			} else {
				http.Redirect(w, r, createCustomerViewError, http.StatusFound)
			}
		}
	}
}

//UpdateCustomerAccountPage UpdateCustomerAccountPage
func (h *Six910Handler) UpdateCustomerAccountPage(w http.ResponseWriter, r *http.Request) {
	ccuuss, suc := h.getUserSession(r)
	h.Log.Debug("session suc", ccuuss)
	if suc {
		if h.isStoreCustomerLoggedIn(ccuuss) {
			//cupvars := mux.Vars(r)
			//ccuemail := cupvars["email"]
			//var uname string
			//ccuemail := ccuuss.Values["username"]
			//h.Log.Debug("ccuemail: ", ccuemail)
			//if ccuemail != nil {
			//uname = ccuemail.(string)
			//}

			hd := h.getHeader(ccuuss)
			//cc.CustomerAccount.Customer.ID
			cc := h.getCustomerCart(ccuuss)
			cus := h.API.GetCustomerID(cc.CustomerAccount.Customer.ID, hd)
			addlst := h.API.GetAddressList(cus.ID, hd)
			//cus := h.API.GetCustomer(uname, hd)
			h.Log.Debug("cus: ", *cus)
			h.Log.Debug("cus ID: ", h.getCustomerID(ccuuss))
			var ucaocp CustomerPage
			ml := h.MenuService.GetMenuList()
			h.getCartTotal(ccuuss, ml, hd)
			ucaocp.MenuList = ml
			ucaocp.Customer = cus
			ucaocp.AddressList = addlst
			ucaocp.StateList = h.StateService.GetStateList("states")
			h.Templates.ExecuteTemplate(w, customerUpdatePage, &ucaocp)
		} else {
			http.Redirect(w, r, customerLoginView, http.StatusFound)
		}
	}
}

//UpdateCustomerAccount UpdateCustomerAccount
func (h *Six910Handler) UpdateCustomerAccount(w http.ResponseWriter, r *http.Request) {
	ccuuuss, suc := h.getUserSession(r)
	h.Log.Debug("session suc", ccuuuss)
	if suc {
		if h.isStoreCustomerLoggedIn(ccuuuss) {
			hd := h.getHeader(ccuuuss)
			uemail := r.FormValue("email")
			ufcus := h.API.GetCustomer(uemail, hd)
			h.Log.Debug("uemail: ", uemail)

			ufirstName := r.FormValue("firstName")
			ulastName := r.FormValue("lastName")
			ucompany := r.FormValue("company")
			ucity := r.FormValue("city")
			ustate := r.FormValue("state")
			uzip := r.FormValue("zip")
			uphone := r.FormValue("phone")
			h.Log.Debug("ufcus: ", ufcus)
			var success bool
			if ufcus != nil {
				ufcus.City = ucity
				ufcus.Company = ucompany
				ufcus.FirstName = ufirstName
				ufcus.LastName = ulastName
				ufcus.Phone = uphone
				ufcus.State = ustate
				ufcus.Zip = uzip
				res := h.API.UpdateCustomer(ufcus, hd)
				success = res.Success
			}
			if success {
				http.Redirect(w, r, customerInfoView, http.StatusFound)
			} else {
				http.Redirect(w, r, customerInfoViewFail, http.StatusFound)
			}
		} else {
			http.Redirect(w, r, customerLoginView, http.StatusFound)
		}
	}
}

//CustomerAddAddressPage CustomerAddAddressPage
func (h *Six910Handler) CustomerAddAddressPage(w http.ResponseWriter, r *http.Request) {
	acauss, suc := h.getUserSession(r)
	h.Log.Debug("session suc", acauss)
	if suc {
		if h.isStoreCustomerLoggedIn(acauss) {
			var uname string
			caddemail := acauss.Values["username"]
			if caddemail != nil {
				uname = caddemail.(string)
			}
			h.Log.Debug("caddemail: ", caddemail)
			hd := h.getHeader(acauss)
			cus := h.API.GetCustomer(uname, hd)
			h.Log.Debug("cus: ", cus)
			h.Templates.ExecuteTemplate(w, customerCreateAddressPage, &cus)
		} else {
			http.Redirect(w, r, customerLoginView, http.StatusFound)
		}
	}
}

//CustomerAddAddress CustomerAddAddress
func (h *Six910Handler) CustomerAddAddress(w http.ResponseWriter, r *http.Request) {
	caaass, suc := h.getUserSession(r)
	h.Log.Debug("session suc", caaass)
	if suc {
		if h.isStoreCustomerLoggedIn(caaass) {
			hd := h.getHeader(caaass)
			idstr := r.FormValue("cid")
			id, _ := strconv.ParseInt(idstr, 10, 64)
			aadmail := r.FormValue("email")
			ufaacus := h.API.GetCustomer(aadmail, hd)
			h.Log.Debug("aadmail: ", aadmail)

			h.Log.Debug("ufaacus: ", ufaacus)
			var success bool
			if ufaacus != nil && ufaacus.ID == id {
				address := r.FormValue("address")
				city := r.FormValue("city")
				state := r.FormValue("state")
				zip := r.FormValue("zip")
				country := r.FormValue("country")
				atype := r.FormValue("type")
				var nad sdbi.Address
				nad.Address = address
				nad.City = city
				nad.State = state
				nad.Zip = zip
				nad.Country = country
				nad.Type = atype
				nad.CustomerID = ufaacus.ID
				res := h.API.AddAddress(&nad, hd)
				success = res.Success
			}
			if success {
				http.Redirect(w, r, customerInfoView, http.StatusFound)
			} else {
				http.Redirect(w, r, customerInfoViewFail, http.StatusFound)
			}
		} else {
			http.Redirect(w, r, customerLoginView, http.StatusFound)
		}
	}
}

//DeleteCustomerAddress DeleteCustomerAddress
func (h *Six910Handler) DeleteCustomerAddress(w http.ResponseWriter, r *http.Request) {
	cdaass, suc := h.getUserSession(r)
	h.Log.Debug("session suc", cdaass)
	if suc {
		if h.isStoreCustomerLoggedIn(cdaass) {
			hd := h.getHeader(cdaass)
			iddastr := r.FormValue("id")
			id, _ := strconv.ParseInt(iddastr, 10, 64)
			ciddastr := r.FormValue("cid")
			cid, _ := strconv.ParseInt(ciddastr, 10, 64)
			dares := h.API.DeleteAddress(id, cid, hd)
			h.Log.Debug("dares: ", dares)
			if dares.Success {
				http.Redirect(w, r, customerInfoView, http.StatusFound)
			} else {
				http.Redirect(w, r, customerInfoViewFail, http.StatusFound)
			}
		} else {
			http.Redirect(w, r, customerLoginView, http.StatusFound)
		}
	}
}
