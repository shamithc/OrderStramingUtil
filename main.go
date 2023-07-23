package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"coinome.in/OrderStramingUtil/database"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// album represents data about a record album.
type open_trade_order_buckets struct {
	OrderType string `json:"order_type"`
	Market    string `json:"market"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
}

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.GET("/testRedis", testRedis)
	router.POST("/updateOrderBook", updateOrderBook)
	router.GET("/fetchOrderBook/:market", fetchOrderBook)

	router.Run("localhost:8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop through the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// getAlbums responds with the list of all albums as JSON.
func testRedis(c *gin.Context) {
	fmt.Println("Hello World!")
	const key string = "players"
	dbClient := database.CreateClient()
	defer dbClient.Close()

	// Sorted set with Values
	z := redis.Z{
		Score:  11,
		Member: "member",
	}

	_, err := dbClient.ZAdd(database.Ctx, "players", z).Result()

	// Hash
	_, err = dbClient.HSet(database.Ctx, "sample-sum", "BTC-INR-100", 10).Result()

	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, "")
}

// getAlbums responds with the list of all albums as JSON.
func fetchOrderBook(c *gin.Context) {
	market := c.Param("market")
	var string_builder strings.Builder

	string_builder.WriteString("buy-")
	string_builder.WriteString(market)
	sorted_buy_key := string_builder.String()

	string_builder.Reset()
	string_builder.WriteString("sell-")
	string_builder.WriteString(market)
	sorted_sell_key := string_builder.String()

	fmt.Println(market)

	dbClient := database.CreateClient()
	defer dbClient.Close()

	buyCmd := dbClient.ZRevRangeWithScores(database.Ctx, sorted_buy_key, 0, 25)

	if err := buyCmd.Err(); err != nil {
		panic(err)
	}

	var buySlice [][]string

	for _, z := range buyCmd.Val() {
		var priceSlice []string
		string_builder.Reset()
		price := int(z.Score)
		string_builder.WriteString(market)
		string_builder.WriteString("-")
		string_builder.WriteString(strconv.Itoa(price))
		hash_field := string_builder.String()
		string_builder.Reset()
		string_builder.WriteString("buy")
		string_builder.WriteString("-sum")
		hash_key := string_builder.String()

		quantity, err := dbClient.HGet(database.Ctx, hash_key, hash_field).Result()
		if err != nil {
			panic(err)
		}

		priceSlice = append(priceSlice, strconv.Itoa(price))
		priceSlice = append(priceSlice, quantity)

		buySlice = append(buySlice, priceSlice)

	}

	sellCmd := dbClient.ZRangeWithScores(database.Ctx, sorted_sell_key, 0, 25)

	if err := sellCmd.Err(); err != nil {
		panic(err)
	}

	var sellSlice [][]string

	for _, z := range sellCmd.Val() {
		var priceSlice []string
		string_builder.Reset()
		price := int(z.Score)
		string_builder.WriteString(market)
		string_builder.WriteString("-")
		string_builder.WriteString(strconv.Itoa(price))
		hash_field := string_builder.String()
		string_builder.Reset()
		string_builder.WriteString("sell")
		string_builder.WriteString("-sum")
		hash_key := string_builder.String()

		quantity, err := dbClient.HGet(database.Ctx, hash_key, hash_field).Result()
		if err != nil {
			panic(err)
		}

		priceSlice = append(priceSlice, strconv.Itoa(price))
		priceSlice = append(priceSlice, quantity)

		sellSlice = append(sellSlice, priceSlice)

	}

	fmt.Println(buySlice)
	fmt.Println(sellSlice)

	c.IndentedJSON(http.StatusOK, gin.H{"buy": buySlice, "sell": sellSlice})
}

const sorted_key_buy_prefix string = "buy"
const sorted_key_sell_prefix string = "sell"
const hash_buy_sum string = "buy_sum"
const hash_sell_sum string = "sell_sum"

// Handle delete if quantity is 0
func updateOrderBook(c *gin.Context) {
	var string_builder strings.Builder
	var openTradeOrderBucket open_trade_order_buckets

	fmt.Println(openTradeOrderBucket)
	if err := c.BindJSON(&openTradeOrderBucket); err != nil {
		panic(err)
		return
	}

	dbClient := database.CreateClient()
	defer dbClient.Close()

	string_builder.WriteString(openTradeOrderBucket.OrderType)
	string_builder.WriteString("-")
	string_builder.WriteString(openTradeOrderBucket.Market)

	sorted_key := string_builder.String()
	string_builder.WriteString("-")
	string_builder.WriteString(strconv.Itoa(openTradeOrderBucket.Price))

	sorted_member := string_builder.String()

	z := redis.Z{
		Score:  float64(openTradeOrderBucket.Price),
		Member: sorted_member,
	}

	_, err := dbClient.ZAdd(database.Ctx, sorted_key, z).Result()

	if err != nil {
		panic(err)
	}

	string_builder.Reset()

	string_builder.WriteString(openTradeOrderBucket.OrderType)
	string_builder.WriteString("-sum")
	hash_key := string_builder.String()
	string_builder.Reset()
	string_builder.WriteString(openTradeOrderBucket.Market)
	string_builder.WriteString("-")
	string_builder.WriteString(strconv.Itoa(openTradeOrderBucket.Price))
	hash_field := string_builder.String()

	_, err = dbClient.HSet(database.Ctx, hash_key, hash_field, openTradeOrderBucket.Quantity).Result()

	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, "Done")
}
