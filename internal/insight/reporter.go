package insight

import (
	"context"
	"log"
	"time"

	pb "github.com/Yuuk111/Go-NetGate/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 定义一个接口，抽象出日志发送的细节，这样我们就可以在未来轻松替换不同的日志发送实现（例如 gRPC、HTTP、文件等）
type LogSender interface {
	SendLog(item *pb.LogItem)
	Close() error
}

type gRPCReporter struct {
	conn   *grpc.ClientConn
	client pb.LogAnalyserClient
	logCh  chan *pb.LogItem
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGRPCReporter(targetAddr string, bufferSize int) (LogSender, error) {
	// 建立 gRPC 连接，目前使用 insecure 连接
	conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewLogAnalyserClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	reporter := &gRPCReporter{
		conn:   conn,
		client: client,
		logCh:  make(chan *pb.LogItem, bufferSize), //使用带缓冲的通道，避免发送日志时阻塞主流程
		ctx:    ctx,
		cancel: cancel,
	}

	// 启动独立的后台消费协程
	go reporter.startWorker()

	return reporter, nil
}

// SendLog 供网关业务/中间件调用的 API
func (r *gRPCReporter) SendLog(item *pb.LogItem) {
	select {
	case r.logCh <- item:

	default:
		//如果通道已满，丢弃日志并记录警告，避免阻塞主流程
		log.Printf("⚠️ [Reporter] 日志通道已满，丢弃日志: %v", item)
	}
}

func (r *gRPCReporter) startWorker() {
	var err error
	var stream pb.LogAnalyser_StreamLogsClient

	for {
		select {
		case <-r.ctx.Done():
			if stream != nil {
				_, _ = stream.CloseAndRecv() //关闭流，等待服务器确认
			}
			log.Println("✅ [Insight] 后台上报 Worker 已安全退出")
			return

		case item := <-r.logCh:
			// 断线重连
			if stream == nil {
				// 如果流不存在，尝试建立新的流连接
				stream, err = r.client.StreamLogs(r.ctx)
				if err != nil {
					log.Printf("❌ [Reporter] 无法建立 gRPC 流连接，正在重试: %v ，当前日志被丢弃", err)
					time.Sleep(1 * time.Second)
					continue
				}
				log.Println("✅ [Insight] 与 AI Agent 的 gRPC 流连接建立成功")
			}

			// 发送日志到服务器
			err = stream.Send(item)
			if err != nil {
				log.Printf("❌ [Reporter] 发送日志失败，流可能已断开：%v ", err)
				stream = nil //将流置空，触发下次循环时重连
			}
		}
	}
}

func (r *gRPCReporter) Close() error {
	r.cancel()            //通知 worker 协程退出
	close(r.logCh)        //关闭日志通道，释放资源
	return r.conn.Close() //关闭 gRPC 连接
}
