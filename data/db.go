package data

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
	"time"
)

/* REDIS "SCHEMA"
This keeps track of the tokens
tokens:<user_token>  =>  <last used time>

This keeps track of user data, which are key/value mappings
userdata:<user_token>:<key>  =>  <value>

Right now there is no expiration date set on anything. If we were to expire tokens, we would want to
also delete their userdata. Moving forward we'll run a periodic job to look for tokens whose last
used time was > a month or so, and then remove all data of that token at once.
*/

const PreTokens = "tokens:"     // Prefix for any registered token, used or unused
const PreUserdata = "userdata:" // Prefix for any user token:key/value mapping

type RedisDB struct {
	userData redis.Conn
	appData  redis.Conn
}

func (r *RedisDB) InitDB() {
	dbPort := os.Getenv("DBPORT")
	conn, err := redis.Dial("tcp", "localhost:" + dbPort)
	if err != nil {
		log.Fatal(err)
	}
	r.userData = conn
	r.appData = conn
}

func (r RedisDB) RegisterToken(newToken string) error {
	key := PreTokens + newToken
	_, err := r.appData.Do("SET", key, time.Now())
	return err
}

func (r RedisDB) Set(token, key string, val interface{}) error {
	// Set the last used time of the token
	inuseKey := PreTokens + token
	_, err := r.appData.Do("SET", inuseKey, time.Now())
	if !checkErr(err) {
		return err
	}

	// Set the actual data
	dbKey := PreUserdata + token + ":" + key
	_, err = r.userData.Do("SET", dbKey, val)
	return err
}

func (r RedisDB) Get(token, key string) (error, string) {
	dbKey := PreUserdata + token + ":" + key
	ret, err := r.userData.Do("GET", dbKey)
	if !checkErr(err) {
		return err, ""
	}
	str, err := redis.String(ret, nil)
	return err, str
}

func (r RedisDB) CheckIfTokenExists(key string) bool {
	ret, err := redis.Bool(r.appData.Do("EXISTS", PreTokens+ key))
	checkErr(err)

	return ret
}

/*
Returns false if there was an error, otherwise true.
*/
func checkErr(err error) bool {
	if err != nil {
		// Do more logging here
		fmt.Println(fmt.Errorf("Error: %s\n", err.Error()))
	}
	return err == nil
}
