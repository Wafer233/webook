# Webook
## Preface

### Introduction
This is a readnote like application, with the basic function of user authentication, content publishing, basic interactive.

### Main Tech Stack
Go, Gin, GORM, MYSQL, Redis, JWT, Kafka(Sarama)

### Development
Now webook is a monolithic application, with `DDD` principle and `RESTful` APIs.

While the future development will mainly focus on migrating to a `microservices` architecture by `gRPC` and adding features such as a search function using `Elasticsearch`.

## Feature
### DDD
#### monolithic

    webook/
    ├── internal/                   # Application core based on DDD principles
    │   ├── domain/                 # Domain layer: Core business entities and logic
    │   │   └── ...                 # User, Article, and Interactive models
    │   ├── repository/             # Repository layer: Data access abstractions
    │   │   ├── ...                 # Repository interfaces and implementations
    │   │   ├── cache/              # Infrastructure: Cache implementations
    │   │   │   └── ...
    │   │   └── dao/                # Infrastructure: Database persistence
    │   │       └── ...
    │   └── service/                # Application layer: Business use cases
    │       └── ...                 # Orchestrates domain operations

#### microservices

`TODO`

Seperate `articles`, `users` , and `interactive` .

### RESTful APIs
#### Source
A resource is the data we want to work with, such as `users`, `articles`.
Each resource is represented by a URL.
    
    pub := group.Group("/pub")

	pub.GET("/:id", handler.PubDetail) //different aid with different url
#### HTTP Methods
    GET //Read data

    POST //Create new data

    PUT //Update data

    DELETE  //Delete data

#### Stateless
Each request must carry all the necessary information 

    tokenHeader := ctx.GetHeader("Authorization")


#### Response
The server usually returns data in JSON format.


    ctx.JSON(http.StatusOK, Result{
                //...
            })

### Database
#### GORM
`GORM` is an `ORM`(Object-Relational Mapping) library for Go. Here deal with `MYSQL`, and basic sql `CRUD`.

    type User struct {
        ID   uint
        Name string
    }

    // This means: SELECT * FROM users WHERE id = 1
    var user User
    db.First(&user, 1) 

#### DAO
`DAO` (Data Access Object) is the layer in code architecture that handles all the database operations. `GORM` used here.
#### Cache
`Cache` (like Redis) is used to store frequently accessed data temporarily, to avoid hitting the database every time. Here use `go-redis`.


### CORS

`CORS` stands for Cross-Origin Resource Sharing. It’s a security feature in the browser that prevents front-end code from calling APIs on a different domain (unless the server allows it).

A middleware deal with different `front-end` and `back-end`


    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"x-jwt-token"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }))


### JWT login
#### Ver.1 Cookie + session（With State）

	sess := sessions.Default(ctx)
	sess.Set("user_id", user.Id)
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	sess.Save()

#### Ver.2 JWT (Stateless)

    now := time.Now()
    if claims.ExpiresAt.Sub(now) < time.Second*50 {

        claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 1))
        tokenStr, err = token.SignedString([]byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"))
        if err != nil {
            log.Print("jwt signing error:", err)
        }

        ctx.Header("x-jwt-token", tokenStr)

    }
    ctx.Set("claims", claims)


### Kubernetes

#### Workflow
1. Containerize application, `redis`, `MYSQL`, `webook`

2. Create a deployment.yaml file to describe:

3. Create a service.yaml file to expose app 

4. Optional. Create pv.yaml and pvc.yaml to make Persistent Volume (database)

5. Deploy Kubernetes cluster
```
    kubectl apply -f deployment.yaml
    kubectl apply -f service.yaml
```


### Wire

#### DI (Dependency Injection)
If there are many dependencies and the hierarchy is complex, `initialization` will become very messy!


#### Workflow
1. Write wire.go
    //go:build wireinject
    wire.NewSet( 
        //...
    )

    func InitFunc() {
        wire.Build{
            //...
        }
    }

2. Use code `wire` in terminal
3. Get `wire_gen.go`

### Kafka

Apache Kafka is a high-performance, distributed message queue system.

It’s commonly used for:
1. Communication between microservices

2. Real-time logging and event streaming

3. Decoupling producers and consumers

#### Workflow

Introducing Kafka to decouple `InteactiveService`
```
// producer 
val, err := json.Marshal(evt)
if err != nil {
    return err
}
_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
    Topic: TopicReadEvent,
    Value: sarama.StringEncoder(val),
})
return err
```
```
// consumer

ctx, cancel := context.WithTimeout(context.Background(), time.Second)

defer cancel()

return i.repo.IncrReadCnt(ctx, "article", event.Aid)
```




