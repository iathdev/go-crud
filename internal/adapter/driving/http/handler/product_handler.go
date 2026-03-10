package handler

import (
	"errors"
	"learning-go/internal/adapter/driving/http/response"
	"learning-go/internal/application/dto"
	"learning-go/internal/application/port/input"
	domainerror "learning-go/internal/domain/error"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productCommand input.ProductCommandPort
	productQuery   input.ProductQueryPort
}

func NewProductHandler(productCommand input.ProductCommandPort, productQuery input.ProductQueryPort) *ProductHandler {
	return &ProductHandler{
		productCommand: productCommand,
		productQuery:   productQuery,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	res, err := h.productCommand.CreateProduct(c.Request.Context(), req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, res)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	res, err := h.productQuery.GetProduct(c.Request.Context(), id)
	if err != nil {
		handleProductError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination = dto.PaginationRequest{Page: 1, PageSize: 10}
	}

	res, err := h.productQuery.ListProducts(c.Request.Context(), pagination)
	if err != nil {
		response.InternalServerError(c, "")
		return
	}

	response.SuccessWithMetadata(c, http.StatusOK, res.Items, res.Metadata)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	res, err := h.productCommand.UpdateProduct(c.Request.Context(), id, req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.productCommand.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		handleProductError(c, err)
		return
	}

	response.Success(c, http.StatusOK, nil, "common.deleted_successfully")
}

func handleProductError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerror.ErrInvalidInput):
		response.BadRequest(c, "common.bad_request")
	case errors.Is(err, domainerror.ErrNotFound):
		response.NotFound(c, "product.not_found")
	default:
		response.InternalServerError(c, "")
	}
}
