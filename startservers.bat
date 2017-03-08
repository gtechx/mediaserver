set ip=192.168.1.50
set scip=192.168.1.50
start servercenter\servercenter.exe -ip=%ip%
start mediareceiveserver\mediareceiveserver.exe -ip=%ip% -scip=%scip%
start broadcastserver\broadcastserver.exe -ip=%ip% -scip=%scip%