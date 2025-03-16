from sys import getsizeof

from core.analysis.graph import DependencyGraph, dependency_graph


class FormulaLoader: # если будет слишком много функций связи с бд или различных стратегий тогда отделить в отдельного менеджера grpc
    """
    Выступает в роли прослойки между графом с формулами и базой данных, основная цель - загрузка значений в активный пул
    Берет на себя:
        связь с базой данных
        надстройку над графом
        связь по grpc с сервером рассылки

    """

    def __init__(self, graph: DependencyGraph):
        self.graph = graph

    def load_formulas_from_db(self, formulas_from_db):
        """
        Загружает формулы из БД и строит зависимости. необходимо оптимизировать в графе метод загрузки
        """

    def trigger_updates(self, formula_ids):
        # тут скорее всего надо делать через паттерн состояния или стратегии, но пока на похер хоть как надо сделать
        """должен по gRpc отправлять данные на Go сервер"""

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


manager = FormulaLoader(graph=dependency_graph)
