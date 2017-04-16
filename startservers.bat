call killservers.bat
set ip=192.168.1.50
set scip=192.168.1.50
start servercenter\servercenter.exe -ip=%ip%
ping localhost -n 2 > nul
start broadcastserver\broadcastserver.exe -ip=%ip% -scip=%scip%
ping localhost -n 2 > nul
start mediareceiveserver\mediareceiveserver.exe -ip=%ip% -scip=%scip%