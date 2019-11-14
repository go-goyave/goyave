# Controllers

[[toc]]

## Defining controllers

Controllers are files containing a collection of Handlers related to a specific feature. Each feature should have its own package. For example, if you have a controller handling user registration, user profiles, etc, you should create a `http/controllers/user` package. Creating a package for each feature has the advantage of cleaning up route definitions a lot and helps keeping a clean structure for your project.

Let's take a very simple CRUD as an example for a controller definition:
**http/controllers/product/product.go**:
``` go
func Store(response *goyave.Response, request *goyave.Request) {
    product := model.Product{
        Name: request.Data["name"].(string),
        Price: request.Data["price"].(float64),
    }
    database.GetConnection().Create(&product)
    response.Status(http.StatusCreated)
}

func Show(response *goyave.Response, request *goyave.Request) {
    product := model.Product{}
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    if database.GetConnection().First(&product, id).RecordNotFound() {
        response.Status(http.StatusNotFound)
    } else {
        response.JSON(http.StatusOK, product)
    }
}

func Update(response *goyave.Response, request *goyave.Request) {
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    product := model.Product{}
    db := database.GetConnection()
    if db.Select("id").First(&product, id).RecordNotFound() {
        response.Status(http.StatusNotFound)
    } else {
        db.Model(&product).Update("name", request.Data["name"].(string))
    }
}

func Destroy(response *goyave.Response, request *goyave.Request) {
    id, _ := strconv.ParseUint(request.Params["id"], 10, 64)
    product := model.Product{}
    db := database.GetConnection()
    if db.Select("id").First(&product, id).RecordNotFound() {
        response.Status(http.StatusNotFound)
    } else {
        db.Delete(&product)
    }
}
```

::: tip
Learn how to handle database errors [here](https://gorm.io/docs/error_handling.html).
:::

## Handlers

A `Handler` is a `func(*goyave.Response, *goyave.Request)`. The first parameter lets you write a response, and the second contains all the information extracted from the raw incoming request.

Read about the available request information in the [Requests](./requests) section.

Controller handlers contain the business logic of your application. They should be concise and focused on what matters for this particular feature in your application. For example, if you develop a service manipulating images, the image processing code shouldn't be written in controller handlers. In that case, the controller handler would simply pass the correct parameters to your image processor and write a response.

``` go
// This handler receives an image, optimizes it and sends the result back.
func OptimizeImage(response *goyave.Response, request *goyave.Request) {
    files := request.Data["image"].([]filesystem.File)
    optimizedImg := processing.OptimizeImage(files[0])
    response.Write(http.StatusOK, optimizedImg)
}
```
::: tip
Setting the `Content-Type` header is not necessary. `response.Write` automatically detects the content type and sets the header accordingly, if the latter has not been defined already.
:::

## Naming conventions

- Controller packages are named after the model they are mostly using, in a singular form. For example a controller for a `Product` model would be called `http/controllers/product`. If a controller isn't related to a model, then give it an expressive name.
- Controller handlers are always **exported** so they can be used when registering routes. All functions which aren't handlers **must be unexported**.
- CRUD operations naming and routing:

| Method           | URI                  | Handler name | Description                    |
|------------------|----------------------|--------------|--------------------------------|
| `GET`            | `/product`           | `Index()`    | Get the products list          |
| `GET`            | `/product/create`    | `Create()`   | Show the product creation form |
| `POST`           | `/product`           | `Store()`    | Create a product               |
| `GET`            | `/product/{id}`      | `Show()`     | Show a product                 |
| `GET`            | `/product/{id}/edit` | `Edit()`     | Show the product update form   |
| `PUT` or `PATCH` | `/product/{id}`      | `Update()`   | Update a product               |
| `DELETE`         | `/product/{id}`      | `Destroy()`  | Delete a product               |
