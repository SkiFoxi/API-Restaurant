package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	neturl "net/url"
	"time"

	"restaurant-demo/internal/models"
)

type CoreClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewCoreClient(baseURL string, logger *slog.Logger) *CoreClient {
	return &CoreClient{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

func (c *CoreClient) RegisterRestaurant(rest models.RestaurantRegistration) (int64, error) {
	c.logger.Debug("Registering restaurant", "name", rest.Name)
	data, err := json.Marshal(rest)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Post(c.baseURL+"/internal/restaurants", "application/json", bytes.NewReader(data))
	if err != nil {
		c.logger.Error("Failed to register restaurant", "error", err)
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("unexpected status: %d", resp.StatusCode)
		c.logger.Error("Registration failed", "status", resp.StatusCode)
		return 0, err
	}
	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("Failed to decode response", "error", err)
		return 0, err
	}
	c.logger.Info("Restaurant registered", "id", result.ID)
	return result.ID, nil
}

// FindRestaurantByName ищет ресторан по точному имени.
// Возвращает (0, nil), если ресторан с таким именем ещё не зарегистрирован —
// это ожидаемый случай (например, самый первый запуск), а не ошибка.
func (c *CoreClient) FindRestaurantByName(name string) (int64, error) {
	c.logger.Debug("Looking up restaurant by name", "name", name)
	url := fmt.Sprintf("%s/internal/restaurants?name=%s", c.baseURL, neturl.QueryEscape(name))
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.logger.Error("Failed to look up restaurant", "error", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.logger.Debug("Restaurant not found by name, will register a new one", "name", name)
		return 0, nil
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status: %d", resp.StatusCode)
		c.logger.Error("Restaurant lookup failed", "status", resp.StatusCode)
		return 0, err
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("Failed to decode restaurant lookup response", "error", err)
		return 0, err
	}
	c.logger.Info("Found existing restaurant", "id", result.ID, "name", name)
	return result.ID, nil
}

func (c *CoreClient) CreateMenuItem(item models.MenuItem) error {
	c.logger.Debug("Creating menu item", "name", item.Name)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Post(c.baseURL+"/internal/menu", "application/json", bytes.NewReader(data))
	if err != nil {
		c.logger.Error("Failed to create menu item", "error", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("unexpected status: %d", resp.StatusCode)
		c.logger.Error("Create menu item failed", "status", resp.StatusCode)
		return err
	}
	c.logger.Info("Menu item created", "name", item.Name)
	return nil
}
//Получаем список новых заказов 
func (c *CoreClient) FetchNewOrders(restaurantID int64) ([]models.Order, error) {
	c.logger.Debug("Fetching new orders", "restaurant_id", restaurantID)
	url := fmt.Sprintf("%s/internal/orders?restaurant_id=%d&status=new", c.baseURL, restaurantID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.logger.Error("Failed to fetch orders", "error", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status: %d", resp.StatusCode)
		c.logger.Error("Fetch orders failed", "status", resp.StatusCode)
		return nil, err
	}
	var respData models.OrderListResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		c.logger.Error("Failed to decode orders", "error", err)
		return nil, err
	}
	if len(respData.Items) > 0 {
		c.logger.Info("Fetched new orders", "count", len(respData.Items))
	}
	return respData.Items, nil
}

//Обновление статуса заказа
func (c *CoreClient) UpdateOrderStatus(orderID int64, status string) error {
	c.logger.Debug("Updating order status", "order_id", orderID, "status", status)
	payload := map[string]string{"status": status}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/internal/orders/%d/status", c.baseURL, orderID)
	//формируем тело запроса с методом patch
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	//Добавляем заголовок в тело запроса
	req.Header.Set("Content-Type", "application/json")
	//Выполняем запрос на основной API (DO универсальный запрос с созданным телом)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to update order status", "error", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status: %d", resp.StatusCode)
		c.logger.Error("Update status failed", "status", resp.StatusCode)
		return err
	}
	c.logger.Info("Order status updated", "order_id", orderID, "status", status)
	return nil
}