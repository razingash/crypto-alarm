import asyncio

from apps.analytics.crud import get_actual_components, get_needed_fields_from_endpoint, add_trigger_history
from apps.analytics.gateway_router import send_triggered_formulas
from apps.binance.external_router import BinanceAPI
from core.analysis.graph import dependency_graph
from core.logger import custom_logger


class BinanceAPIOrchestrator:
    """
    запускает периодические задачи
    на первый взгляд выглядит не плохо, вот только большое количество тасок он сейчас не вывезет
    """
    def __init__(self, binance_api: BinanceAPI):
        self.binance_api = binance_api
        self.is_binance_online = True
        self.tasks: dict[str, asyncio.Task] = {} # лучше это не менять
        self.task_cooldowns: dict[str, int] = {}

    async def start(self):
        """Запуск фоновых задач. первая задача должна быть проверка апи"""
        response = await self.binance_api.get_apiv3_accessibility()
        await self.check_binance_response(response)

        if self.is_binance_online:
            await self.launch_needed_api()

    async def launch_needed_api(self) -> None:
        """выбирает апи, которые нужны для получения актуальных данных в граф, и убирает ненужные"""
        print("launching needed API tasks...")
        data = await get_actual_components()
        current_apis = set(data.keys())
        running_apis = set(self.tasks.keys()) - {"weights"}

        outdated_apis = running_apis - current_apis
        for api in outdated_apis:
            print(f"stopping outdated API: {api}")

            task = self.tasks.pop(api, None)
            self.task_cooldowns.pop(api, None)

            if task and not task.done():
                task.cancel()
            else:
                custom_logger.log_with_path(
                    level=3,
                    msg=f"{type(task), task, self.tasks}",
                    filename="StrangeError.log"
                )

        for api, (cooldown, formulas_amount) in data.items():
            if formulas_amount <= 0:
                continue

            existing_task = self.tasks.get(api)
            if existing_task and not existing_task.done():
                continue

            print(f"Starting API task: {api} (cooldown: {cooldown}seconds)")
            self.launch_api(api, cooldown)

    async def adjust_api_task_cooldown(self, api: str, new_cooldown: int) -> None:
        """Updates the frequency of proxification for a particular API, if necessary"""
        if api not in self.tasks:
            return

        old_cooldown = self.task_cooldowns.get(api)
        if old_cooldown == new_cooldown:
            return

        task = self.tasks.pop(api)
        self.task_cooldowns.pop(api)
        if not task.done():
            task.cancel()
        self.launch_api(api, new_cooldown)

    def launch_api(self, api: str, cooldown: int) -> None:
        if api == "/v3/ticker/price":
            self.tasks[api] = asyncio.create_task(self.update_ticker_current_price(cooldown=cooldown))
            self.task_cooldowns[api] = cooldown
        elif api == "/v3/ticker/24hr":
            self.tasks[api] = asyncio.create_task(self.update_price_change_24h(cooldown=cooldown))
            self.task_cooldowns[api] = cooldown
        elif api == "/v3/ping":
            self.tasks[api] = asyncio.create_task(self.check_api_status_loop(cooldown=cooldown))
            self.task_cooldowns[api] = cooldown

    async def check_binance_response(self, response):
        """checks Binacne availability"""
        if not isinstance(response, dict) and not isinstance(response, list) or response is True: # бинанс лег
            if self.is_binance_online is True:  # первое срабатывание - документация
                custom_logger.log_with_path(
                    level=3,
                    msg=f"Binance DOWN",
                    filename="BinanceAvailablity.log"
                )
                self.is_binance_online = False
            if len(self.tasks) > 1:
                for task in self.tasks:
                    task.cancel()
                self.launch_api("/v3/ping", 60)
        elif self.is_binance_online is False:  # бинанс ожил
            custom_logger.log_with_path(
                level=3,
                msg=f"Binance UP",
                filename="BinanceAvailablity.log"
            )
            self.is_binance_online = True
            for task in self.tasks:
                if isinstance(task, asyncio.Task):
                    task.cancel()
            await self.launch_needed_api()

    async def check_api_status_loop(self, cooldown: int):
        """Проверка доступности API."""
        while True:
            await asyncio.sleep(cooldown)
            response = await self.binance_api.get_apiv3_accessibility()
            await self.check_binance_response(response)
            if self.is_binance_online is True:
                break

    async def update_ticker_current_price(self, cooldown: int):
        while True:
            print('update_ticker_current_price')
            await asyncio.sleep(cooldown)
            response = await self.binance_api.get_ticker_current_price()
            await self.check_binance_response(response)
            currencies = await get_needed_fields_from_endpoint(endpoint="/v3/ticker/price")

            data_for_graph = extract_data_from_ticker_current_price(response, currencies)
            triggered_formulas = dependency_graph.update_variables_topological_Kahn(data_for_graph)
            if len(triggered_formulas) > 0:
                result, variable_values = dependency_graph.get_formulas_variables(triggered_formulas)
                await add_trigger_history(result, variable_values)

                await send_triggered_formulas(formulas=triggered_formulas)

    async def update_price_change_24h(self, cooldown: int):
        while True:
            print('update_price_change_24h')
            await asyncio.sleep(cooldown)
            response = await self.binance_api.get_price_change_24h()
            await self.check_binance_response(response)
            fields = await get_needed_fields_from_endpoint(endpoint="/v3/ticker/24hr")

            data_for_graph = extract_data_from_price_change_24h(response, fields)
            triggered_formulas = dependency_graph.update_variables_topological_Kahn(data_for_graph)
            if len(triggered_formulas) > 0:
                result, variable_values = dependency_graph.get_formulas_variables(triggered_formulas)
                await add_trigger_history(result, variable_values)

                await send_triggered_formulas(formulas=triggered_formulas)


def extract_data_from_price_change_24h(dataset: list, fields: dict[str, list]) -> dict:
    dataset_dict = {data["symbol"]: data for data in dataset}
    result = {}

    for symbol, field_list in fields.items():
        if symbol in dataset_dict:
            data = dataset_dict[symbol]
            for field in field_list:
                if field in data:
                    result[f"{symbol}_{field}"] = float(data[field])

    return result

def extract_data_from_ticker_current_price(dataset: list, currencies: dict[str, list[str]]) -> dict:
    """тут только symbol и price"""
    dataset_dict = {data["symbol"]: data for data in dataset}
    result = {}

    for symbol in currencies.keys():
        data = dataset_dict.get(symbol)
        if data:
            result[f"{symbol}_price"] = float(data["price"])

    return result
