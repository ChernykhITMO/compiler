# easyLang

## Типы данных
- int
- float
- string
- bool
- void
- char
- T[] — массив элементов типа `T` (например, `int[]`, `float[]`, `int[][]`)

## Операторы
- if
- else
- while
- for
- return
- break
- continue

## Синтаксис

```
int a = 5
string s = "abc"

function sum(int a, int b) int {
    int c = a + b
    retrun c
}

while (a < b) {
  a = a - 1
}

for (int i = 0; i < 10; i = i + 1) {}
```

### Массивы
Объявление
```
int[] arr
int[][] matrix
```
Выделение памяти
```
arr = new int[5]
matrix = new int[3]
matrix[0] = new int[4]
```

Индексация и присваивание
```
arr[0] = 10
arr[1] = arr[0] + 5

int i = 0
while (i < 5) {
    arr[i] = i * 2
    i = i + 1
}

int sum = 0
i = 0
while (i < 5) {
    sum = sum + arr[i]
    i = i + 1
}
```

