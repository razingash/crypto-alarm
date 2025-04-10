from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession

from apps.analytics.crud import get_formula_by_id
from core.analysis.manager import formula_manager
from db.postgre import postgres_db

router = APIRouter()

"""
message xchange with crypto-gateway service. implemented through Reast and not GRPC,
since in the future the project will be converted into SaaS
"""

@router.post(path='/formula/{pk}/')
async def add_formula(pk: int, session: AsyncSession = Depends(postgres_db.session_dependency)):
    """adds the formula to the graph"""
    formula = await get_formula_by_id(session, pk)
    res = formula_manager.add_formulas_to_graph(formula, pk)
    if not res:
        return {"error": res}


@router.delete(path='/formula/{pk}/')
async def remove_formula(pk: int):
    """удаляет формулу из графа"""
    res = formula_manager.remove_formulas_from_graph(pk)
    if res is not True:
        return {"error": res}


@router.put(path='/formula/{pk}/')
async def update_formula(pk: int, session: AsyncSession = Depends(postgres_db.session_dependency)):
    """обновляет формулу в графе"""
    formula = await get_formula_by_id(session, pk)
    res = formula_manager.update_formula_in_graph(formula, pk)

    if not res:
        return {"error": res}
