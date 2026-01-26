package handlers

import (
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type BrandHandler struct {
	brandService *services.BrandService
}

func NewBrandHandler(brandService *services.BrandService) *BrandHandler {
	return &BrandHandler{brandService: brandService}
}

func (h *BrandHandler) ListBrands(c *gin.Context) {
	var query services.BrandListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	normalizedQuery := c.Request.URL.Query()
	normalizedQuery.Set("include_inactive", strconv.FormatBool(query.IncludeInactive))
	normalizedQuery.Set("with_specs", strconv.FormatBool(query.WithSpecs))
	cacheKey := cache.NamespaceKeyWithQuery(cache.NamespaceBrandList, normalizedQuery)
	if entry, ok := tryServeCached(c, cacheKey); ok {
		if entry != nil {
			response.Success(c, entry.Data)
		}
		return
	}

	brands, err := h.brandService.ListBrands(&query)
	if err != nil {
		response.InternalError(c, "获取品牌列表失败")
		return
	}
	saveCachedResponse(c, cacheKey, brands, cacheTTLBrandList)
	response.Success(c, brands)
}

func (h *BrandHandler) GetBrand(c *gin.Context) {
	brandID := c.Param("id")
	cacheKey := cache.NamespaceKey(cache.NamespaceBrandDetail, brandID)
	if entry, ok := tryServeCached(c, cacheKey); ok {
		if entry != nil {
			response.Success(c, entry.Data)
		}
		return
	}
	brand, err := h.brandService.GetBrand(brandID)
	if err != nil {
		if err.Error() == "brand not found" {
			response.NotFound(c, "品牌不存在")
			return
		}
		response.InternalError(c, "获取品牌失败")
		return
	}
	saveCachedResponse(c, cacheKey, brand, cacheTTLBrandDetail)
	response.Success(c, brand)
}

func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var req services.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.brandService.CreateBrand(&req)
	if err != nil {
		response.InternalError(c, "创建品牌失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	response.Created(c, brand)
}

func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	brandID := c.Param("id")
	var req services.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.brandService.UpdateBrand(brandID, &req)
	if err != nil {
		if err.Error() == "brand not found" {
			response.NotFound(c, "品牌不存在")
			return
		}
		response.InternalError(c, "更新品牌失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	response.Success(c, brand)
}

func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	brandID := c.Param("id")
	if err := h.brandService.DeleteBrand(brandID); err != nil {
		if err.Error() == "brand not found" {
			response.NotFound(c, "品牌不存在")
			return
		}
		response.InternalError(c, "删除品牌失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	cache.BumpNamespace(cache.NamespaceBrandSpecs)
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *BrandHandler) ListBrandSpecs(c *gin.Context) {
	brandID := c.Param("id")
	includeInactive := c.Query("include_inactive") == "true"
	normalizedQuery := c.Request.URL.Query()
	normalizedQuery.Set("brand_id", brandID)
	normalizedQuery.Set("include_inactive", strconv.FormatBool(includeInactive))
	cacheKey := cache.NamespaceKeyWithQuery(cache.NamespaceBrandSpecs, normalizedQuery)
	if entry, ok := tryServeCached(c, cacheKey); ok {
		if entry != nil {
			response.Success(c, entry.Data)
		}
		return
	}
	specs, err := h.brandService.ListBrandSpecs(brandID, includeInactive)
	if err != nil {
		response.InternalError(c, "获取规范失败")
		return
	}
	saveCachedResponse(c, cacheKey, specs, cacheTTLBrandSpecs)
	response.Success(c, specs)
}

func (h *BrandHandler) CreateBrandSpec(c *gin.Context) {
	brandID := c.Param("id")
	var req services.CreateBrandSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	spec, err := h.brandService.CreateBrandSpec(brandID, &req)
	if err != nil {
		if err.Error() == "brand not found" {
			response.NotFound(c, "品牌不存在")
			return
		}
		response.InternalError(c, "创建规范失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	cache.BumpNamespace(cache.NamespaceBrandSpecs)
	response.Created(c, spec)
}

func (h *BrandHandler) UpdateBrandSpec(c *gin.Context) {
	specID := c.Param("specId")
	var req services.UpdateBrandSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	spec, err := h.brandService.UpdateBrandSpec(specID, &req)
	if err != nil {
		if err.Error() == "brand spec not found" {
			response.NotFound(c, "规范不存在")
			return
		}
		response.InternalError(c, "更新规范失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	cache.BumpNamespace(cache.NamespaceBrandSpecs)
	response.Success(c, spec)
}

func (h *BrandHandler) DeleteBrandSpec(c *gin.Context) {
	specID := c.Param("specId")
	if err := h.brandService.DeleteBrandSpec(specID); err != nil {
		if err.Error() == "brand spec not found" {
			response.NotFound(c, "规范不存在")
			return
		}
		response.InternalError(c, "删除规范失败")
		return
	}
	cache.BumpNamespace(cache.NamespaceBrandList)
	cache.BumpNamespace(cache.NamespaceBrandDetail)
	cache.BumpNamespace(cache.NamespaceBrandSpecs)
	response.Success(c, gin.H{"message": "删除成功"})
}
