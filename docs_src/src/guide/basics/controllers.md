---
meta:
  - name: "og:title"
    content: "Controllers - Goyave"
  - name: "twitter:title"
    content: "Controllers - Goyave"
  - name: "title"
    content: "Controllers - Goyave"
---

# Controllers

[[toc]]

## Defining controllers

Controllers are files containing a collection of Handlers related to a specific feature. Each feature should have its own package. For example, if you have a controller handling user registration, user profiles, etc, you should create a `http/controller/user` package. Creating a package for each feature has the advantage of cleaning up route definitions a lot and helps keeping a clean structure for your project.

Let's take a very simple CRUD as an example for a controller definition:
**http/controller/product/product.go**:
``` go
func Index(response *goyave.Response, request *goyave.Request) {
	products := []model.Product{}
	result := database.Conn().Find(&products)
	if response.HandleDatabaseError(result) {
		response.JSON(http.StatusOK, products)
	}
}

func Show(response *goyave.Response, request *goyave.Request) {
	product := model.Product{}
	result := database.Conn().First(&product, request.Params["id"])
	if response.HandleDatabaseError(result) {
		response.JSON(http.StatusOK, product)
	}
}

func Store(response *goyave.Response, request *goyave.Request) {
	product := model.Product{
		Name:  request.String("name"),
		Price: request.Numeric("price"),
	}
	if err := database.Conn().Create(&product).Error; err != nil {
		response.Error(err)
	} else {
		response.JSON(http.StatusCreated, map[string]uint{"id": product.ID})
	}
}

func Update(response *goyave.Response, request *goyave.Request) {
	product := model.Product{}
	db := database.Conn()
	result := db.Select("id").First(&product, request.Params["id"])
	if response.HandleDatabaseError(result) {
		if err := db.Model(&product).Update("name", request.String("name")).Error; err != nil {
			response.Error(err)
		}
	}
}

func Destroy(response *goyave.Response, request *goyave.Request) {
	product := model.Product{}
	db := database.Conn()
	result := db.Select("id").First(&product, request.Params["id"])
	if response.HandleDatabaseError(result) {
		if err := db.Delete(&product).Error; err != nil {
			response.Error(err)
		}
	}
}
```

::: tip
- Learn how to handle database errors [here](https://gorm.io/docs/error_handling.html).
- It is not necessary to add `response.Status(http.StatusNoContent)` at the end of `Update` and `Destroy` because the framework automatically sets the response status to 204 if its body is empty and no status has been set.
:::

## Handlers

A `Handler` is a `func(*goyave.Response, *goyave.Request)`. The first parameter lets you write a response, and the second contains all the information extracted from the raw incoming request.

Read about the available request information in the [Requests](./requests.html) section.

Controller handlers contain the business logic of your application. They should be concise and focused on what matters for this particular feature in your application. For example, if you develop a service manipulating images, the image processing code shouldn't be written in controller handlers. In that case, the controller handler would simply pass the correct parameters to your image processor and write a response.

``` go
// This handler receives an image, optimizes it and sends the result back.
func OptimizeImage(response *goyave.Response, request *goyave.Request) {
    optimizedImg := processing.OptimizeImage(request.File("image")[0])
    response.Write(http.StatusOK, optimizedImg)
}
```
::: tip
Setting the `Content-Type` header is not necessary. `response.Write` automatically detects the content type and sets the header accordingly, if the latter has not been defined already.
:::

## Naming conventions

- Controller packages are named after the model they are mostly using, in a singular form. For example a controller for a `Product` model would be called `http/controller/product`. If a controller isn't related to a model, then give it an expressive name.
- Controller handlers are always **exported** so they can be used when registering routes. All functions which aren't handlers **must be unexported**.
- CRUD operations naming and routing:

| Method           | URI             | Handler name | Description           |
|------------------|-----------------|--------------|-----------------------|
| `GET`            | `/product`      | `Index()`    | Get the products list |
| `POST`           | `/product`      | `Store()`    | Create a product      |
| `GET`            | `/product/{id}` | `Show()`     | Show a product        |
| `PUT` or `PATCH` | `/product/{id}` | `Update()`   | Update a product      |
| `DELETE`         | `/product/{id}` | `Destroy()`  | Delete a product      |
