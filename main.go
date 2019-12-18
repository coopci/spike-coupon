package main

import (
	"encoding/json"
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

// curl -XPOST -d'id=12312&quatity=5000' http://localhost:8081/addCoupon
func addCoupon(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.Form["id"][0]
	var quatity = r.Form["quatity"][0]
	var key = "coupon::" + id
	// fmt.Printf("key = %s %T\n", key, key)
	// fmt.Printf("quatity = %s %T\n", quatity, quatity)
	// fmt.Printf("redisClient=%p \n", redisClient)
	delta, err := strconv.ParseInt(quatity, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "quatity must be an integer: %s!", quatity)
	}
	incrby, err := redisClient.HIncrBy(key, "left", delta).Result()

	// for key, value := range r.Form {
	// 	fmt.Printf("%s = %s \n", key, value)
	// }
	fmt.Fprintf(w, "%s.left set to %d", key, incrby)
}

// curl -XPOST -d'id=12312&per_user_max=5' http://localhost:8081/setPerUserMax
func setPerUserMax(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.Form["id"][0]
	var perUserMax = r.Form["per_user_max"][0]
	var key = "coupon::" + id

	perUserMaxInt64, err := strconv.ParseInt(perUserMax, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "per_user_max must be an integer: %s!", perUserMax)
	}

	result, err := redisClient.HSet(key, "per_user_max", perUserMaxInt64).Result()
	_ = result
	// for key, value := range r.Form {
	// 	fmt.Printf("%s = %s \n", key, value)
	// }
	fmt.Fprintf(w, "%s.per_user_max set to %d\n", key, perUserMaxInt64)
}

// curl  http://localhost:8081/getInfo?id=12312
func getInfo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var id = r.URL.Query()["id"][0]

	var key = "coupon::" + id

	result, err := redisClient.HGetAll(key).Result()
	_ = err
	b, err := json.Marshal(result)
	fmt.Fprintf(w, string(b))
}

//申请优惠券
// curl -XPOST -d'id=12312&uid=zhang3&quatity=1' http://localhost:8081/apply
// curl -XPOST -d'id=12312&uid=11&quatity=10' http://localhost:8081/apply
// ab -p ab.txt -T application/x-www-form-urlencoded -c 64 -n 10000 -k http://localhost:8081/apply
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
	if perUserMax == "" {
		fmt.Fprintf(w, "No such coupon: %s", id)
		return
	}
	perUserMaxInt64, err := strconv.ParseInt(perUserMax, 10, 64)
	if err != nil {
		// redis 里存的数据有问题
		fmt.Fprintf(w, "Internal error, %s.per_user_max is %s", key, perUserMax)
		return
	}
	if quatityInt64 > perUserMaxInt64 {
		fmt.Fprintf(w, "%s.per_user_max set is %d, this user %s is asking for %d", key, perUserMaxInt64, uid, quatityInt64)
	}

	// coupon::12312,uid::zhang3
	userKey := key + ",uid::" + uid
	newQuatityInt64, err := redisClient.HIncrBy(userKey, "granted", quatityInt64).Result()
	// newQuatityInt64, err := strconv.ParseInt(newQuatity, 10, 64)

	if newQuatityInt64 > perUserMaxInt64 {
		// 超了每用户的最大申请数，需要退回
		newQuatityInt64, err := redisClient.HIncrBy(userKey, "granted", -quatityInt64).Result()
		_ = err
		fmt.Fprintf(w, "%s.per_user_max set is %d, this user %s already has %d", key, perUserMaxInt64, uid, newQuatityInt64)
		return
	}
	// 减库存
	newLeftInt64, err := redisClient.HIncrBy(key, "left", -quatityInt64).Result()
	if newLeftInt64 < 0 {
		// 库存减完了，需要退回。
		newLeftInt64, err := redisClient.HIncrBy(key, "left", quatityInt64).Result()
		_ = err
		newQuatityInt64, err := redisClient.HIncrBy(userKey, "granted", -quatityInt64).Result()
		_ = newQuatityInt64
		fmt.Fprintf(w, "%s.left has only %d.", key, newLeftInt64)
	} else {
		fmt.Fprintf(w, "Granted %d. User %s now has %d.\n", quatityInt64, uid, newQuatityInt64)
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
	http.HandleFunc("/getInfo", getInfo)
	http.HandleFunc("/apply", apply)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
