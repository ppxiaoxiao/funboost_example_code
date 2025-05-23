import time
import json
from funboost import boost, BrokerEnum, BoosterParams, PriorityConsumingControlConfig

# 修改函数定义，使用RabbitMQ
@boost(BoosterParams(
    queue_name="first_queue_priority", 
    broker_kind=BrokerEnum.RABBITMQ_AMQPSTORM,  # 改为使用RabbitMQ
    concurrent_num=10,
    is_show_message_get_from_broker=True,
    broker_exclusive_config={'x-max-priority': 10, 'durable': True, 'no_ack': False}  # 设置持久化和消息确认
))
def first_processor(payload, priority):  # 修改这里，直接接受payload和priority参数
    """
    处理第一个队列的消息
    payload: 消息的负载部分
    priority: 消息的优先级
    """
    print(f"接收到first_queue_priority消息 - 优先级: {priority}, 负载: {payload}")
    
    try:
        # 处理逻辑...
        processed_result = {"original_payload": payload, "processed_data": "处理结果"}
        
        # 发送到第二个队列
        second_processor.publish(
            {"data": {
                "payload": processed_result,
                "priority": priority,
                "timestamp": time.time()
            }},
            priority_control_config=PriorityConsumingControlConfig(other_extra_params={'priroty': priority})
        )
        
        return "第一阶段处理完成"
    except Exception as e:
        print(f"处理消息时出错: {e}")
        import traceback
        traceback.print_exc()
        raise

# 修改第二个函数定义
@boost(BoosterParams(
    queue_name="second_queue_priority", 
    broker_kind=BrokerEnum.RABBITMQ_AMQPSTORM,  # 改为使用RabbitMQ
    concurrent_num=5,
    is_show_message_get_from_broker=True,
    broker_exclusive_config={'x-max-priority': 10, 'durable': True, 'no_ack': False}  # 设置持久化和消息确认
))
def second_processor(data):  # 保持这个函数不变，因为它接收的是通过Python push的消息
    """
    处理第二个队列的消息
    """
    print(f"接收到second_queue_priority消息: {data}")
    try:
        priority = data.get("priority", 5)
        payload = data.get("payload", {})
        
        # 最终处理逻辑...
        print(f"最终处理结果，优先级: {priority}, 数据: {payload}")
        
        return "第二阶段处理完成"
    except Exception as e:
        print(f"处理消息时出错: {e}")
        import traceback
        traceback.print_exc()
        raise

if __name__ == "__main__":
    print("使用funboost_config.py中配置的RabbitMQ连接")
    
    # 启动消费者
    print("启动消费者进程...")
    first_processor.consume()
    second_processor.consume()
    
    # 手动发布测试消息
    try:
        print("发布测试消息...")
        # 发布优先级为1的消息
        first_processor.publish(
            {"payload": {"test": "这是一条低优先级测试消息"}, "priority": 1},
            priority_control_config=PriorityConsumingControlConfig(other_extra_params={'priroty': 1})
        )
        
        # 发布优先级为5的消息
        first_processor.publish(
            {"payload": {"test": "这是一条中优先级测试消息"}, "priority": 5},
            priority_control_config=PriorityConsumingControlConfig(other_extra_params={'priroty': 5})
        )
        
        # 发布优先级为9的消息
        first_processor.publish(
            {"payload": {"test": "这是一条高优先级测试消息"}, "priority": 9},
            priority_control_config=PriorityConsumingControlConfig(other_extra_params={'priroty': 9})
        )
        
        print("测试消息发布成功")
    except Exception as e:
        print(f"发布测试消息失败: {e}")
    
    # 保持程序运行
    while True:
        time.sleep(10)
