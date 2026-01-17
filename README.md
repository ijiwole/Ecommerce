# E-Commerce Backend API

A robust, scalable e-commerce backend API built with Go (Golang) and MongoDB. This RESTful API provides complete functionality for user management, product catalog, shopping cart, order processing, and address management.

## 🚀 Features

### User Management
- User registration and authentication
- JWT-based authentication with refresh tokens
- Admin user management
- Secure password hashing with bcrypt

### Product Management
- Admin can create products
- Public product listing with pagination
- Product search functionality
- Automatic filtering of sold products

### Shopping Cart
- Add/remove products from cart
- View cart contents
- Cart checkout with payment method selection
- Instant buy functionality
- Idempotency protection (10-second window)

### Order Management
- Order creation and tracking
- Support for multiple payment methods (Digital/Cash on Delivery)
- Order history per user
- Prevents duplicate orders

### Address Management
- Add multiple addresses
- Edit home and work addresses
- Delete addresses
- Retrieve all user addresses

## 🛠️ Tech Stack

- **Language**: Go 1.25.1
- **Web Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt
- **Validation**: go-playground/validator
- **Environment Variables**: godotenv

## 📋 Prerequisites

- Go 1.25.1 or higher
- MongoDB (local or remote)
- Git

## 🔧 Installation

1. **Clone the repository**
   ```bash
   git clone <github.com/ijiwole>
   cd ecommerce-yt
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up MongoDB**
   
   Option 1: Using Docker Compose (Recommended)
   ```bash
   docker-compose up -d
   ```
   
   Option 2: Install MongoDB locally
   - Follow MongoDB installation guide for your OS
   - Ensure MongoDB is running on `localhost:27017`

4. **Configure environment variables**
   
   Create a `.env` file in the root directory:
   ```env
   SECRET_KEY=your-secret-key-here-min-32-chars
   MONGODB_URL=mongodb://localhost:27017
   PORT=8000
   ```

5. **Run the application**

   Option 1: With Hot Reload (Recommended for Development)
   ```bash
   # Install air (if not already installed)
   go install github.com/air-verse/air@latest
   
   # If 'air' command is not found, add Go bin to your PATH:
   # For zsh/bash, add this to ~/.zshrc or ~/.bashrc:
   # export PATH=$PATH:$(go env GOPATH)/bin
   # Then run: source ~/.zshrc (or source ~/.bashrc)
   
   # Run with hot reload
   air
   ```
   
   Option 2: Standard Run
   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:8000` (or the port specified in `.env`)
   
   **Note**: With `air`, the application will automatically rebuild and restart whenever you save changes to any `.go` file.

## 📁 Project Structure

```
ecommerce-yt/
├── controllers/          # HTTP request handlers
│   ├── controllers.go   # User & admin authentication, product management
│   ├── cart.go          # Cart operations
│   └── address.go       # Address management
├── database/            # Database operations
│   ├── database-setup.go # MongoDB connection
│   ├── cart.go          # Cart database operations
│   └── address.go       # Address database operations
├── models/              # Data models
│   └── models.go        # User, Product, Order, Address models
├── routes/              # API route definitions
│   └── routes.go        # Route configuration
├── middleware/          # Middleware functions
│   └── middleware.go    # Authentication & authorization
├── tokens/              # JWT token management
│   └── tokengen.go      # Token generation & validation
├── helpers/             # Utility functions
│   ├── response.go      # Standardized API responses
│   └── pagination.go    # Pagination helpers
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── docker-compose.yaml  # MongoDB Docker setup
└── .env                 # Environment variables (create this)
```

## 🔐 API Endpoints

### Public Endpoints (No Authentication)

#### User Authentication
- `POST /api/v1/users/signup` - User registration
- `POST /api/v1/users/login` - User login
- `POST /api/v1/admin/signup` - Admin registration
- `POST /api/v1/admin/login` - Admin login

#### Product Search
- `GET /api/v1/users/productview?search=<query>` - Search products by name
- `GET /api/v1/users/search?name=<name>&category=<category>` - Advanced product search

### Protected Endpoints (Requires Authentication)

#### Products
- `GET /api/v1/products?page=1&page_size=10` - Get all products (paginated)

#### Cart Operations
- `POST /api/v1/cart/add?id=<product_id>` - Add product to cart
- `DELETE /api/v1/cart/remove?id=<product_id>` - Remove product from cart
- `GET /api/v1/cart` - Get cart items
- `POST /api/v1/cart/checkout` - Checkout cart
  - Body (optional): `{"digital": true, "cod": false}`
- `POST /api/v1/cart/instantbuy?id=<product_id>` - Instant buy
  - Body (optional): `{"digital": true, "cod": false}`

#### Address Management
- `GET /api/v1/address` - Get all addresses
- `POST /api/v1/address` - Add new address
- `PUT /api/v1/address/home` - Update home address
- `PUT /api/v1/address/work` - Update work address
- `DELETE /api/v1/address?id=<address_id>` - Delete address

### Admin Endpoints (Requires Admin Authentication)

- `POST /api/v1/admin/addproduct` - Create new product

## 📝 API Usage Examples

### 1. User Registration

```bash
POST /api/v1/users/signup
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "phone": "1234567890",
  "password": "password123"
}
```

### 2. User Login

```bash
POST /api/v1/users/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "success": true,
  "message": "Logged In Successfully",
  "user_id": "user_id_here",
  "token": "jwt_token_here",
  "refresh_token": "refresh_token_here"
}
```

### 3. Add Product to Cart

```bash
POST /api/v1/cart/add?id=<product_id>
Header: token: <your_jwt_token>
```

### 4. Checkout Cart

```bash
POST /api/v1/cart/checkout
Header: token: <your_jwt_token>
Content-Type: application/json

{
  "digital": true,
  "cod": false
}
```

### 5. Create Product (Admin)

```bash
POST /api/v1/admin/addproduct
Header: token: <admin_jwt_token>
Content-Type: application/json

{
  "product_name": "Laptop",
  "price": 999,
  "rating": 5,
  "image": "https://example.com/image.jpg"
}
```

## 🔒 Authentication

All protected endpoints require a JWT token in the request header:

```
Header: token: <your_jwt_token>
```

Or using Bearer token format:
```
Header: Authorization: Bearer <your_jwt_token>
```

## 🎯 Key Features Explained

### Idempotency
- Checkout operations are idempotent within a 10-second window
- Duplicate checkout requests within 10 seconds return the same order ID
- Prevents accidental duplicate orders from network retries

### Sold Product Filtering
- Products that have been sold are automatically excluded from product listings
- Prevents purchasing already-sold items

### Payment Methods
- **Digital**: Credit card, PayPal, etc.
- **COD**: Cash on Delivery
- Defaults to COD if not specified

### Pagination
- Product listings support pagination
- Query parameters: `page` (default: 1) and `page_size` (default: 10, max: 100)

## 🐳 Docker Setup

The project includes a `docker-compose.yaml` file for easy MongoDB setup:

```bash
docker-compose up -d
```

This starts MongoDB on port 27017 with:
- Username: `root`
- Password: `rootpassword`

## 🧪 Testing

You can test the API using tools like:
- Postman
- cURL
- HTTPie
- Any REST client

Example with cURL:

```bash
# Login
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Get Products (with token)
curl -X GET http://localhost:8000/api/v1/products?page=1&page_size=10 \
  -H "token: <your_token>"
```

## 📦 Dependencies

Key dependencies:
- `github.com/gin-gonic/gin` - Web framework
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/go-playground/validator/v10` - Input validation
- `github.com/joho/godotenv` - Environment variable management

## 🔄 Response Format

All API responses follow a standardized format:

**Success Response:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Error message here"
}
```

**Login Response:**
```json
{
  "success": true,
  "message": "Logged In Successfully",
  "user_id": "user_id",
  "token": "jwt_token",
  "refresh_token": "refresh_token"
}
```

## 🚨 Error Handling

The API uses standard HTTP status codes:
- `200` - Success
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## 📝 Notes

- JWT tokens expire after 24 hours
- Refresh tokens expire after 7 days
- Cart is automatically cleared after successful checkout
- Products are marked as sold after order completion
- Admin users can create products and have elevated privileges

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the MIT License.

## 👤 Author

Built with ❤️ for e-commerce backend development

---

**Happy Coding! 🚀**
