


### About

- Users can create their own expressions. Custom keyboard includes fractions,
powers up, roots and modules. Trigonometric functions will never be supported.

- All cryptocurrencies supported by all available Binance endpoints are accessible — over 3000 in total.
This covers approximately 99.8% of all cryptocurrencies on Binance, excluding only the newest ones not yet fully integrated across all endpoints.

- All variables from all significant endpoints, such as /v3/ticker/24hr, /v3/ticker/price, etc. are available. There is currently no access to websockets

- If for some reason Binance stops supporting cryptocurrencies or variables received from specific APIs,
system will block access to these data and deactivates all related formulas and variables


### Quickstart

1) Database initialization: run following command in crypto-analyzer directory
    ```bash 
    python manage.py initialization
   ```

2) Frontend: run this command in frontend directory to run ReactJs server 
    ```bash 
    npm start
   ```

3) Client's connection with the database: in crypto-gateway/cmd/ run following command to run fiber server
    ```bash 
    go run main.go
   ```

4) Analytics: This service is only responsible for periodic analytics and requests for Baninas.
   run following command in crypto-analyzer directory 
   ```bash 
    python main.py
   ```
