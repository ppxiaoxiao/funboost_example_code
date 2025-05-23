package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

func main() {
    ctx := context.Background()
    
    // 连接Redis
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "redis123",
        DB:       0,
    })
    
    // 测试连接
    pong, err := rdb.Ping(ctx).Result()
    if err != nil {
        fmt.Println("Redis连接失败:", err)
        return
    }
    fmt.Println("Redis连接成功:", pong)
    
    // 准备发送10条消息，优先级不同
    for i := 1; i <= 10; i++ {
        // 计算优先级 (1-10，1为最高优先级)
        priority := i % 10
        if priority == 0 {
            priority = 10
        }
        
        // 创建消息内容
        payload := map[string]interface{}{
            "id":          i,
            "content":     fmt.Sprintf("这是第%d条消息", i),
            "created_at":  time.Now().Format(time.RFC3339),
        }
        
        // 创建data字段
        data := map[string]interface{}{
            "payload":  payload,
            "priority": priority,
        }
        
        // 序列化data字段
        dataBytes, err := json.Marshal(data)
        if err != nil {
            fmt.Println("数据序列化失败:", err)
            continue
        }
        
        // 使用Redis Stream发布消息，直接设置键值对
        streamKey := "first_queue"
        taskID := fmt.Sprintf("task_%d_%d", i, time.Now().UnixNano())
        
        // 重要：添加一个空键，这是funboost期望的
        values := map[string]interface{}{
            "":        string(dataBytes),  // 空键是funboost期望的
            "body":    string(dataBytes),  // 保留body以防需要
            "task_id": taskID,
        }
        
        streamID, err := rdb.XAdd(ctx, &redis.XAddArgs{
            Stream: streamKey,
            ID:     "*", // 自动生成ID
            Values: values,
        }).Result()
        
        if err != nil {
            fmt.Println("消息发送失败:", err)
        } else {
            fmt.Printf("已发送消息 ID: %d, 优先级: %d, Stream ID: %s\n", i, priority, streamID)
        }
        
        // time.Sleep(500 * time.Millisecond)
    }
    
    fmt.Println("所有消息已发送")
}
