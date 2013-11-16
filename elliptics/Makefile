.PHONY: clean

TARGET=test

$(TARGET): libell.so
	go build .

libell.so: logger.o node.o
	g++ -shared -o $@ $^

%.o: %.cpp
	g++ -std=c++0x -fPIC -o $@ -c $^

install:
	install libell.so /usr/lib/libell.so 

clean:
	rm -f *.o *.so *.a $(TARGET)
