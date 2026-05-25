package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Errorf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Errorf("Ошибка при добавлении: %v", err)
		return
	}

	if id == 0 {
		t.Errorf("Ошибка ID не сгенерирован")
	}

	parcel.Number = id

	// get
	storedParcel, err := store.Get(id)
	if err != nil {
		t.Errorf("Ошибка получения посылки: %v", err)
		return
	}

	if storedParcel.Number != parcel.Number || storedParcel.Status != parcel.Status || storedParcel.Address != parcel.Address || storedParcel.Client != parcel.Client {
		t.Errorf("Данные полученной посылки не совпадают")
		return
	}

	err = store.Delete(id)
	if err != nil {
		t.Errorf("Ошибка при удалении посылки: %v", err)
		return
	}
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	_, err = store.Get(id)
	if err == nil {
		t.Errorf("Ошибка: посылка с ID %d все еще существует в БД после удаления", id)
		return
	}
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Errorf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	if err != nil {
		t.Errorf("Ошибка при добавлении посылки: %v", err)
		return
	}

	if id == 0 {
		t.Errorf("ID не сгенерирован")
		return
	}

	parcel.Number = id
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Errorf("Ошибка при обновлении адреса: %v", err)
		return
	}

	storedParcel, err := store.Get(id)
	if err != nil {
		t.Errorf("Ошибка при получении посылки для проверки: %v", err)
		return
	}
	// check
	if storedParcel.Address != newAddress {
		t.Errorf("Адрес не обновился: надо %s, получено %s", newAddress, storedParcel.Address)
		return
	}
	// получите добавленную посылку и убедитесь, что адрес обновился
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Errorf("Ошибка при подключении к базе данных: %v", err)
		return
	}

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Errorf("Ошибка при добавлении посылки")
		return
	}

	if id == 0 {
		t.Errorf("ID не сгенирирован")
		return
	}
	// set status
	newStatus := ParcelStatusRegistered
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Errorf("Ошибка при обновлении статуса: %v", err)
		return
	}
	// обновите статус, убедитесь в отсутствии ошибки
	storedParcel, err := store.Get(id)
	if err != nil {
		t.Errorf("Ошибка при получении посылки для проверки: %v", err)
		return
	}

	if storedParcel.Status != newStatus {
		t.Errorf("Статус не обновился: надо %s, получено %s", newStatus, storedParcel.Status)
	}
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		t.Errorf("Ошибка при подключении к базе данных: %v", err)
		return
	}
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	store := NewParcelStore(db)

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		if err != nil {
			t.Errorf("Ошибка при добавлении адреса: %v", err)
			return
		}

		if id == 0 {
			t.Errorf("ID не сгенирирован")
			return
		}
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Errorf("Ошибка при получении списка: %v", err)
		return
	}
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	if len(storedParcels) != len(parcels) {
		t.Errorf("Ошибка количество полученных посылок не совпадает: надо %d получено %d", len(parcels), len(storedParcels))
		return
	}
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		original, exists := parcelMap[parcel.Number]
		if !exists {
			t.Errorf("Найдена посылка с ID %d, которая не была добавлена", parcel.Number)
			continue
		}

		if original.Client != parcel.Client {
			t.Errorf("Ошибка клиента для посылки %d: надо %d получено %d", parcel.Number, original.Client, parcel.Client)
		}

		if original.Address != parcel.Address {
			t.Errorf("Ошибка адреса для посылки %d: надо %s получено %s", parcel.Number, original.Address, parcel.Address)
		}

		if original.Status != parcel.Status {
			t.Errorf("Ошибка статуса для посылки %d: надо %s получено %s", parcel.Number, original.Status, parcel.Status)
		}
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
