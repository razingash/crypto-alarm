from sys import getsizeof

import httpx
from colorama import Fore, Style

from apps.analytics.crud import get_actual_formulas
from core.analysis.graph import DependencyGraph, dependency_graph
from core.logger import custom_logger
from core.orchestrator import BinanceAPIOrchestrator


class FormulaManager:
    """
    Выступает в роли прослойки между графом с формулами и базой данных, основная цель - загрузка значений в активный пул
    Берет на себя:
        связь с базой данных
        надстройку над графом
        связь по REST с сервером рассылки. в grpc нет смысла
    """
    def __init__(self, graph: DependencyGraph):
        self.graph = graph
        self.client = httpx.AsyncClient(timeout=10)
        self._loaded = False
        self.orchestrator = None

    def set_orchestrator(self, orchestrator: BinanceAPIOrchestrator):
        self.orchestrator = orchestrator

    async def load(self):
        self._loaded = True
        await self.load_formulas_from_db()

    async def send_triggered_formulas(self, formulas_id: list, is_shutted_off: bool):
        """
        отправляет список id формул которые сработали, или которые были отключены из-за различных факторов,
        чтобы сделать рассылку пуш уведомлений"
        """
        data = {"formulas": formulas_id, "is_shutted_off": is_shutted_off}
        try:
            await self.client.post(url="http://localhost:8001/api/v1/triggers/push-notifications", params=data)
        except httpx.ConnectError:
            custom_logger.log_with_path(
                level=2,
                msg=f"Connection error during sending a message to crypto-gateway server, most-likely it was overloaded or shutted-off",
                filename="ExternalErrors.log"
            )

    async def load_formulas_from_db(self):
        """Загружает формулы из БД и строит зависимости. необходимо оптимизировать в графе метод загрузки"""
        formulas = await get_actual_formulas()
        print(Fore.LIGHTYELLOW_EX + 'Trying to load formulas from a database...' + Style.RESET_ALL)

        errors = []
        error = ""
        for row in formulas:
            formula_id = row[0]
            formula = row[1]

            res = self.graph.add_formula(formula, formula_id)
            if res is not True:
                error = res
                errors.append(formula_id)

        if len(errors) > 0:
            print(Fore.LIGHTRED_EX + f"An errors occured during loading formulas, more info in logs/Initialization.log:  {error}" + Style.RESET_ALL)
            custom_logger.log_with_path(
                level=1,
                msg=f"During the formulas loading into the graph through FormulaManager - load_formulas_from_db an error occurred",
                filename="Initialization.log"
            )
        else:
            print(Fore.LIGHTGREEN_EX + f"{len(formulas)} formulas are correctly loaded into the graph {error}" + Style.RESET_ALL)

    async def add_formulas_to_graph(self, formula: str, pk: int):
        res = self.graph.add_formula(formula, pk)
        print(res, self.orchestrator)
        if res is True and self.orchestrator:
            await self.orchestrator.launch_needed_api()
        return res

    async def remove_formulas_from_graph(self, pk: int):
        res = self.graph.remove_formula(pk)
        if res is True and self.orchestrator:
            await self.send_triggered_formulas(formulas_id=[pk], is_shutted_off=False)
            await self.orchestrator.launch_needed_api()
        return res

    async def update_formula_in_graph(self, formula: str, pk: int):
        res = self.graph.remove_formula(pk)
        if res is True and self.orchestrator:
            res = self.graph.add_formula(formula, pk)
            await self.orchestrator.launch_needed_api()

        return res

    def get_graph_full_size(self, obj, seen=None) -> int: # нигде больше не понадобится поэтому оставить тут
        """вычисляет полный размер объекта в байтах"""
        if seen is None:
            seen = set()

        obj_id = id(obj)
        if obj_id in seen:
            return 0
        seen.add(obj_id)
        size = getsizeof(obj)

        if isinstance(obj, dict):
            size += sum(self.get_graph_full_size(k) + self.get_graph_full_size(v) for k, v in obj.items())
        elif isinstance(obj, (list, tuple, set, frozenset)):
            size += sum(self.get_graph_full_size(i) for i in obj)
        elif hasattr(obj, '__dict__'):
            size += self.get_graph_full_size(obj.__dict__)

        print(f"Graph current size is: {size} bytes, or {size / 1_000_000} mb")
        return size

formula_manager = FormulaManager(graph=dependency_graph)
