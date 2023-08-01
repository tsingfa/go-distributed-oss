//消息队列的相关操作
//New 新建消息队列
//Bind 将消息队列与exchange绑定，所有发往exchange的消息会转发到本地消息队列
//Send 往指定消息队列发送消息
//Publish 往指定exchange发送消息
//Consume 生成一个用于接收消息的channel，遍历该ch以获取来自本地队列的消息
//Close 关闭消息队列

package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	channel  *amqp.Channel
	conn     *amqp.Connection
	Name     string
	exchange string
}

// New 创建RabbitMQ结构体指针
func New(s string) *RabbitMQ {
	conn, err := amqp.Dial(s)
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		panic(err)
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	mq.Name = q.Name
	return mq
}

// Bind 将自己的消息队列与一个exchange绑定，
// 所有发往该exchange的消息都可以在自己的消息队列中接收到。
func (q *RabbitMQ) Bind(exchange string) {
	err := q.channel.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)
	if err != nil {
		panic(err)
	}
	q.exchange = exchange
}

// Send 往指定消息队列发消息
func (q *RabbitMQ) Send(queue string, body interface{}) {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	err = q.channel.Publish("",
		queue,
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    bytes,
		})
	if err != nil {
		panic(err)
	}
}

// Publish 往指定exchange发消息
func (q *RabbitMQ) Publish(exchange string, body interface{}) {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	err = q.channel.Publish(exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    bytes,
		})
	if err != nil {
		panic(err)
	}
}

func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	ch, err := q.channel.Consume(q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	return ch
}

func (q *RabbitMQ) Close() {
	_ = q.channel.Close()
	_ = q.conn.Close()
}
