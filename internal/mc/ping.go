package mc

import (
	"context"
	"encoding/json"
	"time"
	"fmt"

	"github.com/Tnze/go-mc/bot"
)


func Ping(ctx context.Context,addr string,timeout time.Duration) (bool,int,time.Duration,error){
	ctx,cancel:=context.WithTimeout(ctx,timeout)
	defer cancel()
	resp ,delay,err:=bot.PingAndListContext(ctx,addr)
	if err!=nil{
		return false,0,0,fmt.Errorf("ping %s: %w",addr,err)
	}
	var status struct{
		Player struct{
			Oline int `json:"online"`
		}`json:"players"`
	}
	if err:=json.Unmarshal(resp,&status);err!=nil{
		return false,0,0,fmt.Errorf("Decode json error:%w",err)
	}
	return true,status.Player.Oline,delay,nil
}