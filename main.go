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
	var per_user_max = r.Form["per_user_max"][0]
	var key = "coupon::" + id
	// fmt.Printf("key = %s %T\n", key, key)
	// fmt.Printf("quatity = %s %T\n", quatity, quatity)
	// fmt.Printf("redisClient=%p \n", redisClient)
	delta, err := strconv.ParseInt(quatity, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "quatity must be an integer: %s!", quatity)
	}
	incrby, err := redisClient.HIncrBy(key, "left", delta).Result()

	perUserMaxInt64, err := strconv.ParseInt(per_user_max, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "per_user_max must be an integer: %s!", per_user_max)
	}

	// for key, value := range r.Form {
	// 	fmt.Printf("%s = %s \n", key, value)
	// }
	fmt.Fprintf(w, "%s.left set to %d", key, incrby)
}

// curl -XPOST -d'id=12312&per_user_max=5' http://localhost:8080/setPerUserMax
func setPerUserMax(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.Form["id"][0]
	var perUserMax = r.Form["per_user_max"][0]
	var key = "coupon::" + id

	perUserMaxInt64, err := strconv.ParseInt(perUserMax, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "per_user_max must be an integer: %s!", perUserMax)
	}

	perUserMaxInt64NewValue, err := redisClient.HSet(key, "per_user_max", perUserMaxInt64).Result()

	// for key, value := range r.Form {
	// 	fmt.Printf("%s = %s \n", key, value)
	// }
	fmt.Fprintf(w, "%s.per_user_max set to %d", key, perUserMaxInt64NewValue)
}

//申请优惠券
// curl -XPOST -d'id=12312&uid=11&quatity=1' http://localhost:8080/apply
func apply(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.Form["id"][0] // 优惠券id
	var quatity = r.Form["quatity"][0]
	var uid = r.Form["uid"][0] // 用户id
	var key = "coupon::" + id
	// fmt.Printf("key = %s %T\n", key, key)
	// fmt.Printf("quatity = %s %T\n", quatity, quatity)
	// fmt.Printf("redisClient=%p \n", redisClient)
	quatityInt64, err := strconv.ParseInt(quatity, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "quatity must be an integer: %s!", quatity)
	}
	perUserMax, err := redisClient.HGet(key, "per_user_max").Result()
	perUserMaxInt64, err := strconv.ParseInt(perUserMax, 10, 64)
	if err != nil {
		// redis 里的存的数据有问题
		fmt.Fprintf(w, "Internal error, %s.per_user_max set is %s", perUserMax)
	}
	if quatityInt64 > perUserMaxInt64 {
		fmt.Fprintf(w, "%s.per_user_max set is %d", perUserMaxInt64)
	}
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
	http.HandleFunc("/setPerUserMax", setPerUserMax)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
