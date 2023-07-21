export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on

OBJ = test_linger
SRC = main.go

all: $(OBJ)
	go mod tidy && go build -gcflags "-N -l" -o ${OBJ} ./$(SRC)

clean:
	rm -fr $(OBJ)

-include .deps
dep:
	echo -n "$(OBJ):" > .deps
	find . -name '*.go' | awk '{print $$0 " \\"}' >> .deps
	echo "" >> .deps