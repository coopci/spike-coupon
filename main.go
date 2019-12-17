package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

// curl -XPOST -d'id=12312&quatity=5000' http://localhost:8080/addCoupon
func addCoupon(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.Form["id"][0]
	var quatity = r.Form["quatity"][0]
	var key = "coupon_left::" + id
	// fmt.Printf("key = %s %T\n", key, key)
	// fmt.Printf("quatity = %s %T\n", quatity, quatity)
	// fmt.Printf("redisClient=%p \n", redisClient)
	delta, err := strconv.ParseInt(quatity, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "quatity must be an integer: %s!", quatity)
	}
	incrby, err := redisClient.IncrBy(key, delta).Result()

	// for key, value := range r.Form {
	// 	fmt.Printf("%s = %s \n", key, value)
	// }
	fmt.Fprintf(w, "%s set to %d", key, incrby)
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// pong, err := redisClient.Ping().Result()
	// fmt.Println(pong, err)
	http.HandleFunc("/", handler)
	http.HandleFunc("/addCoupon", addCoupon)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
