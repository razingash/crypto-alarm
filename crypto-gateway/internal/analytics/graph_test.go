package analytics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFormula(t *testing.T) {
	dg := NewDependencyGraph()

	syntaxTests := []struct {
		id        int
		name      string
		formula   string
		wantError bool
	}{
		{1, "less or equal", "a+b*c-d/(k+c)>=100", false},
		{2, "greater or equal", "a+b*c-d/(k+c)>=100", false},
		{3, "equal", "a+b*c-d/(k+c)==100", false},
		{4, "greater", "a+b*c-d/(k+c)>100", false},
		{6, "power", "a^b>100", false},                  // should work
		{7, "double multiplication", "a**b>100", false}, // also works, but will never be used
		{1, "id already in graph", "a+b*c>=100", true},
		{-1, "negative id", "a+b*c-d/(k+c)<=100", true},
		{0, "unnatural id", "a+b*c-d/(k+c)<=100", true},
		{8, "plus minus '+-'", "a+-b>100", true},
		{9, "double dividing", "a//b+c>100", true},
		{10, "broken expression 1", "(a+b>100", true},
		{11, "broken expression 2", "a+c)>100", true},
		{12, "broken expression 3", "((a+b>100", true},
		{13, "broken expression 4", "a+b))>100", true},
		{14, "broken expression 5", "+a+b>100", true},
		{15, "broken expression 6", "--a+b>100", true},
		{16, "broken expression 7", "*a+b>100", true},
		{17, "broken expression 8", "**a+b>100", true},
		{18, "broken expression 9", "/a+b>100", true},
		{20, "broken expression 10", "a+b+>100", true},
		{21, "broken expression 11", "a+b->100", true},
		{22, "broken expression 12", "a+b/>100", true},
		{23, "broken expression 13", "a+b*>100", true},
		{24, "broken expression 14", "*a+b>100", true},
		{25, "broken expression 15", "a++b>100", true},
		{26, "broken expression 16", "a+b", true},
		{27, "broken expression 17", "a+b<", true},
		{28, "broken expression 18", "a+b>", true},
		{29, "broken expression 19", "a+b<=", true},
		{30, "broken expression 20", "a+b>=", true},
		{31, "broken expression 21", "a+b==", true},
		{32, "broken expression 22", "a+c>--100", true},
		{33, "broken expression 23", "a+c>+100", true},
		{34, "broken expression 24", "a+c>++100", true},
		{35, "broken expression 25", "a+c>*100", true},
		{36, "broken expression 26", "a+c>/100", true},
		{37, "broken expression 27", "a+c>//100", true},
		{38, "broken expression 28", "a+c>**100", true},
	}

	for _, tt := range syntaxTests {
		t.Run(tt.name, func(t *testing.T) {
			err := dg.AddFormula(tt.formula, tt.id)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	formulas := []string{"a+b*c-d/(k+c)>=100", "a+b*c-d/c>=100", "a+b*c-d/k==100", "a+b*c>100", "a^b>100"}

	resultTests := []struct {
		name        string
		setup       func(dg *DependencyGraph)
		expectation func(t *testing.T, dg *DependencyGraph)
	}{
		{
			name: "Graph Result",
			setup: func(dg *DependencyGraph) {
				for i, v := range formulas {
					dg.AddFormula(v, i+1)
				}
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, dg.Graph["a"])
				assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, dg.Graph["b"])
				assert.ElementsMatch(t, []int{1, 2, 3, 4}, dg.Graph["c"])
				assert.ElementsMatch(t, []int{1, 2, 3}, dg.Graph["d"])
				assert.ElementsMatch(t, []int{1, 3}, dg.Graph["k"])
			},
		},
		{
			name: "Graph Formulas and Compiled",
			setup: func(dg *DependencyGraph) {
				for i, v := range formulas {
					dg.AddFormula(v, i+1)
				}
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				for i, v := range formulas {
					assert.Equal(t, v, dg.Formulas[i+1])
					assert.Equal(t, v, dg.Compiled[i+1].String())
				}
			},
		},
		{
			name: "Variables and Cache",
			setup: func(dg *DependencyGraph) {
				for i, v := range formulas {
					dg.AddFormula(v, i+1)
				}
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				assert.ElementsMatch(t, []int{}, dg.Variables)
				assert.ElementsMatch(t, []int{}, dg.Cache)
			},
		},
		{
			name: "Compiled Subexpressions",
			setup: func(dg *DependencyGraph) {
				for i, v := range formulas {
					dg.AddFormula(v, i+1)
				}
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				expectedSubexprs := []string{
					"k+c",
					"(k+c)",
					"d/(k+c)",
					"a+b*c",
					"a+b*c-d/(k+c)",
					"a+b*c-d/(k+c)>=100",
					"d/c",
					"a+b*c-d/c",
					"a+b*c-d/c>=100",
					"d/k",
					"a+b*c-d/k",
					"a+b*c-d/k==100",
					"a+b*c>100",
					"a^b",
					"a^b>100",
					"b*c",
				}

				assert.Equal(t, len(dg.SubexprCompiled), len(expectedSubexprs))
				for _, expr := range expectedSubexprs {
					assert.Contains(t, dg.SubexprCompiled, expr, "missing subexpression: %s", expr)
					assert.NotNil(t, dg.SubexprCompiled[expr], "expression not compiled: %s", expr)
				}
			},
		},
		{
			name: "Subexpressions Weight",
			setup: func(dg *DependencyGraph) {
				for i, v := range formulas {
					dg.AddFormula(v, i+1)
				}
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				expectedWeights := map[string]int{
					"a+b*c":              4,
					"b*c":                4,
					"(k+c)":              1,
					"a+b*c-d/(k+c)":      1,
					"a+b*c-d/(k+c)>=100": 1,
					"a+b*c-d/c":          1,
					"a+b*c-d/c>=100":     1,
					"a+b*c-d/k":          1,
					"a+b*c-d/k==100":     1,
					"a+b*c>100":          1,
					"a^b":                1,
					"a^b>100":            1,
					"d/(k+c)":            1,
					"d/c":                1,
					"d/k":                1,
					"k+c":                1,
				}

				for expr, expectedCount := range expectedWeights {
					t.Run("Weight: "+expr, func(t *testing.T) {
						actualCount, ok := dg.SubexprWeights[expr]
						assert.True(t, ok, "Expected subexpression not found: %s", expr)
						assert.Equal(t, expectedCount, actualCount, "Mismatch in weight for subexpression: %s", expr)
					})
				}
			},
		},
	}

	for _, tt := range resultTests {
		t.Run(tt.name, func(t *testing.T) {
			dg := NewDependencyGraph()
			tt.setup(dg)
			tt.expectation(t, dg)
		})
	}
}

func TestRemoveFormula(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(dg *DependencyGraph)
		expectation func(t *testing.T, dg *DependencyGraph)
	}{
		{
			name: "Single formula",
			setup: func(dg *DependencyGraph) {
				dg.AddFormula("a+b*c-d/(k+c)>=100", 1)
				dg.RemoveFormula(1)
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				assert.Equal(t, NewDependencyGraph(), dg)
			},
		},
		{
			name: "Multiple formulas",
			setup: func(dg *DependencyGraph) {
				dg.AddFormula("a+b*c+d==100", 1)
				dg.AddFormula("b*c==100", 2)
				dg.AddFormula("a+b==100", 3)
				dg.AddFormula("b*c==100", 4)
				dg.UpdateVariablesTopologicalKahn(map[string]float64{"a": 1200, "b": 200, "c": -100, "d": -0.0003})
				dg.RemoveFormula(1)
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				t.Run("Graph", func(t *testing.T) {
					assert.ElementsMatch(t, []int{3}, dg.Graph["a"])
					assert.ElementsMatch(t, []int{2, 3, 4}, dg.Graph["b"])
					assert.ElementsMatch(t, []int{2, 4}, dg.Graph["c"])
				})

				t.Run("Formulas and Compiled", func(t *testing.T) {
					expectedFormulas := map[int]string{
						2: "b*c==100",
						3: "a+b==100",
						4: "b*c==100",
					}
					for id, formula := range expectedFormulas {
						assert.Equal(t, formula, dg.Formulas[id])
						assert.Equal(t, formula, dg.Compiled[id].String())
					}
				})

				t.Run("Cache should be empty", func(t *testing.T) {
					assert.Empty(t, dg.Cache)
				})

				t.Run("Variables", func(t *testing.T) {
					assert.Equal(t, map[string]float64{"a": 1200, "b": 200, "c": -100}, dg.Variables)
				})

				t.Run("SubexprWeights", func(t *testing.T) {
					expectedWeights := map[string]int{
						"a+b":      1,
						"a+b==100": 1,
						"b*c":      2,
						"b*c==100": 2,
					}
					for expr, expectedCount := range expectedWeights {
						actualCount, ok := dg.SubexprWeights[expr]
						assert.True(t, ok, "Expected subexpression not found: %s", expr)
						assert.Equal(t, expectedCount, actualCount, "Mismatch in weight for subexpression: %s", expr)
					}
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dg := NewDependencyGraph()
			tt.setup(dg)
			tt.expectation(t, dg)
			fmt.Println(dg.Graph)
		})
	}
}

func TestUpdateVariablesTopologicalKahn(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(dg *DependencyGraph)
		expectation func(t *testing.T, dg *DependencyGraph)
	}{
		{
			name: "FullTest",
			setup: func(dg *DependencyGraph) {
				dg.AddFormula("a - b == 1000", 1)
				dg.AddFormula("b + c == 100", 2)
				dg.AddFormula("a - b + c < 200", 3)
				dg.AddFormula("ETHBTC_priceChange < 1", 4)
			},
			expectation: func(t *testing.T, dg *DependencyGraph) {
				result := dg.UpdateVariablesTopologicalKahn(map[string]float64{"a": 1200, "b": 200, "c": -100, "ETHBTC_priceChange": -0.0003})

				assert.Equal(t, float64(1200), dg.Variables["a"])
				assert.Equal(t, float64(200), dg.Variables["b"])
				assert.Equal(t, float64(-100), dg.Variables["c"])
				assert.Equal(t, float64(-0.0003), dg.Variables["ETHBTC_priceChange"])
				fmt.Println(dg.Variables)
				assert.ElementsMatch(t, []int{1, 2, 4}, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dg := NewDependencyGraph()
			tt.setup(dg)
			tt.expectation(t, dg)
			fmt.Println(dg.Graph)
		})
	}
}
