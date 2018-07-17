mkdir logs
for /l %%i in (1,1,50) do (start cmd /k "go test ./... %* >logs/test%%i.log & exit /b 0")
