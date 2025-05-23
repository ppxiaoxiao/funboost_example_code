package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/streadway/amqp" // 使用streadway/amqp库连接RabbitMQ
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

func main() {
    // 连接RabbitMQ
    conn, err := amqp.Dial("amqp://admin:admin123@localhost:5672/")
    failOnError(err, "无法连接到RabbitMQ")
    defer conn.Close()

    // 创建通道
    ch, err := conn.Channel()
    failOnError(err, "无法打开通道")
    defer ch.Close()

    // 声明队列，确保队列存在
    queueName := "first_queue_priority"
    args := amqp.Table{
        "x-max-priority": int32(10), // 设置最大优先级为10
        "durable":        true,      // 持久化队列
    }
    
    q, err := ch.QueueDeclare(
        queueName, // 队列名
        true,      // durable 持久化
        false,     // 非自动删除
        false,     // 非独占
        false,     // no-wait
        args,      // 参数
    )
    failOnError(err, "无法声明队列")
    
    fmt.Println("RabbitMQ连接成功，队列已声明:", q.Name)
    
    // 准备发送10条消息，优先级不同
    for i := 1; i <= 10; i++ {
        // 计算优先级 (1-10，10为最高优先级)
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
        
        // 额外参数，包含优先级信息
        extraParams := map[string]interface{}{
            "other_extra_params": map[string]interface{}{
                "priroty": priority, // 注意这里是"priroty"而不是"priority"，这是funboost的命名
            },
        }
        
        // 完整消息结构
        message := map[string]interface{}{
            "payload":  payload,
            "priority": priority, 
            "extra":    extraParams,
        }
        
        // 序列化消息
        messageBytes, err := json.Marshal(message)
        if err != nil {
            fmt.Println("数据序列化失败:", err)
            continue
        }
        
        // 发布消息到RabbitMQ队列
        err = ch.Publish(
            "",        // exchange
            queueName, // routing key (队列名)
            false,     // mandatory
            false,     // immediate
            amqp.Publishing{
                ContentType:  "application/json",
                Body:         messageBytes,
                DeliveryMode: amqp.Persistent, // 持久化消息
                Priority:     uint8(priority), // 设置消息优先级
            })
        
        if err != nil {
            fmt.Println("消息发送失败:", err)
        } else {
            fmt.Printf("已发送消息 ID: %d, 优先级: %d, 队列: %s\n", i, priority, queueName)
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    fmt.Println("所有消息已发送")
}
