from sys import getsizeof

from core.analysis.graph import DependencyGraph, dependency_graph


# сейчас эта фигня не оптимизирована - нужно чтобы в формуле значения вычислялись по частям
class FormulaManager:
    """
    Выступает в роли прослойки между графом с формулами и базой данных
    Берет на себя:
        связь с базой данных
        надстройку над графом
        связь по grpc с сервером рассылки

    """

    def __init__(self, graph: DependencyGraph):
        self.graph = graph

    def load_formulas_from_db(self, formulas_from_db):
        """
        Загружает формулы из БД и строит зависимости.
        """


    def trigger_updates(self, formula_ids):
        """должен по gRpc отправлять данные на Go сервер"""

    def get_graph_full_size(self, obj) -> int: # нигде больше не понадобится поэтому оставить тут
        """вычисляет полный размер объекта в байтах"""
        seen = set()

        obj_id = id(obj)
        if obj_id in seen:
            return 0
        seen.add(obj_id)

        size = getsizeof(obj)

        if isinstance(obj, dict):
            size += sum(self.get_full_size(k) + self.get_full_size(v) for k, v in obj.items())
        elif isinstance(obj, (list, tuple, set, frozenset)):
            size += sum(self.get_full_size(i) for i in obj)
        elif hasattr(obj, '__dict__'):
            size += self.get_full_size(obj.__dict__)

        print(f"Graph current size is: {size} bytes, or {size / 1_000_000} mb")
        return size


manager = FormulaManager(graph=dependency_graph)
