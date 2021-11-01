package main

import (
	"fmt"
	rpio "github.com/stianeikeland/go-rpio/v4"
	"math"
	"os"
	"os/signal"
	"time"
	"github.com/imroc/req"
)


const spiClk = 1000000
const RESET = 0x06
const REF_ = 0x0a
const REF = 0x05
const DATARATE_ = 0x3e
const DATARATE = 0x04
const INPMUX = 0x02
const PGA = 0x03
const RDATA = 0x12

const Pin_start = rpio.Pin(17)
const Pin_drdy =  rpio.Pin(18)

func MUX_(i int	) int {
	return i << 4 | (0x2)
}

func PGA_(i int)int {
	return 1 << 3 | i
}

func int_2_rawvolt(input_int int)int {
	if input_int == 0x7fff {
		return 66666
	}else if input_int == 0x8000{
		return -66666
	}else if input_int > 0x7fff{
		return input_int - 0xffff - 1
	}else {
		return input_int
	}
}

func init()  {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	time.Sleep(time.Second*2)
	pin_start := rpio.Pin(Pin_start)  //这里都是bcm编码，注意
	pin_start.Output()
	pin_start.Low()

	pin_drdy:=rpio.Pin(Pin_drdy)
	pin_drdy.Input()

	//设置spi串口通信
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rpio.SpiSpeed(spiClk)
	rpio.SpiMode(0, 1)

	//重置初始化
	rpio.SpiTransmit(RESET)
	time.Sleep(time.Second*2)
	write_reg(REF, REF_)
	write_reg(DATARATE, DATARATE_)
}

func write_reg(address,reg int)  {
	rpio.SpiTransmit(uint8(1 << 6 | address))
	rpio.SpiTransmit(0x00)
	rpio.SpiTransmit(uint8(reg))
}
func write(data int)  {
	rpio.SpiTransmit(uint8(data))
}

func set_mux(i int)  {
	write_reg(INPMUX,MUX_(i))
}

func set_pga(i int)  {
	write_reg(PGA, PGA_(i))
}

func wait(pin rpio.Pin)  {
	for pin.Read()!=rpio.Low{
	}
}

func read_data_raw(i, pga int) int  {
	set_mux(i)
	set_pga(pga)
	Pin_start.High()
	wait(Pin_drdy)
	write(RDATA)
	a1:=int(rpio.SpiReceive(1)[0])
	a2:=int(rpio.SpiReceive(1)[0])
	b:=(a1 << 8)|(a2)
	Pin_start.Low()
	return b
}

func read_volt(sensor_index int)float64  {
 	var data_useful float64
	for i:=0;i<7;i++{
		data_volt:=int_2_rawvolt(read_data_raw(sensor_index,i))
		if (data_volt != 66666) && (data_volt != -66666){
			data_useful =float64(data_volt)/(math.Pow(2,float64(i)))
		}else {
			break
		}
	}
	return data_useful
}


func main()  {
	// Unmap gpio memory when done
	defer rpio.Close()

	api_url := os.Getenv("api_url")

	sig_chan := make(chan []int, 1)

	go func() {
		for{
			data:=make([]int,6)
			for i:=0;i<6;i++{
				temp:=int(read_volt(i))
				if temp<0 {
					temp=-temp
				}
				data[i]=temp
			}
			sig_chan <- data
			fmt.Println(data)
			println("\n-------------------------------------------")
			time.Sleep(time.Second)
		}
	}()


	post_param := req.Param{}

	go func() {
		for{
			data_:=<-sig_chan
			post_param["sensor0"]=data_[5]
			post_param["sensor1"]=data_[0]
			post_param["sensor2"]=data_[1]
			post_param["sensor3"]=data_[2]
			post_param["sensor4"]=data_[3]
			post_param["sensor5"]=data_[4]
			r,_:=req.Post(api_url,req.BodyJSON(post_param))
			print(r.String())
		}
	}()


	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch

}
