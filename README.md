# Threadly: A Chat-Based CRM System with Business Tools

Threadly is a comprehensive customer relationship management (CRM) system built with Go and Gin, featuring real-time chat capabilities similar to WhatsApp, combined with powerful business management tools.

## 🏗️ Architecture

- **Backend**: Go with Gin framework
- **Frontend**: HTML templates with Tailwind CSS
- **Database**: SQLite (development) / PostgreSQL (production)
- **Authentication**: JWT-based with role-based access
- **Real-time**: WebSocket-like polling for live updates

---

## 🚀 Features of the System

### 🏢 Business Features

#### **Authentication & Security**
- **Secure Login**: Business user authentication with email/password
- **JWT Sessions**: Token-based authentication with automatic renewal
- **Role-based Access**: Business-only routes and permissions
- **Session Management**: Automatic logout and session validation

#### **Dashboard & Analytics**
- **Business Dashboard**: Comprehensive overview with key metrics
- **Revenue Tracking**: Total revenue, order revenue, booking revenue
- **Performance Metrics**: Order counts, booking counts, conversion rates
- **Client Analytics**: Active clients, conversation statistics
- **Visual Charts**: Revenue trends and business performance indicators

#### **Product & Service Management**
- **Product Catalog**: Full CRUD operations for products
- **Service Catalog**: Complete service management system
- **Inventory Tracking**: Stock levels and low-stock alerts
- **Media Upload**: Product images and service photos
- **Pricing Control**: Flexible pricing with discounts and variations

#### **Order Management**
- **Order Processing**: Complete order lifecycle management
- **Status Tracking**: Pending → Confirmed → Fulfilled → Cancelled
- **Order Details**: Line items, quantities, pricing, delivery info
- **Payment Integration**: Payment tracking and partial payment support
- **Order History**: Complete order archive with search and filter

#### **Booking & Scheduling**
- **Service Booking**: Complete booking system for services
- **Calendar Integration**: Date/time scheduling with availability
- **Booking Management**: Status tracking and confirmation system
- **Resource Planning**: Service provider scheduling and capacity
- **Automated Reminders**: Booking confirmations and follow-ups

#### **Client Communication**
- **Real-time Chat**: WhatsApp-like messaging interface
- **Conversation History**: Complete message archive with search
- **Media Sharing**: Image, document, and file sharing in chat
- **Message Types**: Text, images, documents, system notifications
- **Quick Actions**: One-click order creation from chat

#### **Advanced Features**
- **Conversation Progress**: Automated sales funnel tracking
- **Action Management**: Task creation from client messages
- **Quick Actions**: Rapid order/booking creation from chat
- **Payment Requests**: Integrated payment collection system
- **Goal Setting**: Custom business objectives and tracking

---

### 👥 Client Features

#### **Client Authentication**
- **OTP-based Login**: Secure one-time password authentication
- **Email Verification**: Verified client access with email OTP
- **Session Management**: Secure client sessions with tokens
- **Multi-device Support**: Login from multiple devices

#### **Client Dashboard**
- **Personal Profile**: Client information management
- **Order History**: Complete order tracking and status
- **Booking Management**: Service booking overview and calendar
- **Communication Hub**: Centralized messaging interface
- **Business Directory**: Discover and connect with businesses

#### **Client Chat Interface**
- **Rich Messaging**: Text, emojis, media sharing
- **Order Integration**: Direct order placement from chat
- **Booking Integration**: Service booking from conversation
- **Real-time Updates**: Live status updates and notifications
- **Message History**: Searchable conversation archive

#### **Self-Service Features**
- **Order Editing**: Modify order quantities and notes
- **Booking Management**: Reschedule or cancel bookings
- **Payment Processing**: Secure payment integration
- **Profile Updates**: Manage personal information

---

### 🔧 System Features

#### **Database Management**
- **Dual Database Support**: SQLite for development, PostgreSQL for production
- **Auto-migration**: Automatic schema updates and version control
- **Data Integrity**: Foreign key constraints and validation
- **Backup Support**: Database backup and recovery tools

#### **Security & Performance**
- **Input Validation**: Comprehensive data validation and sanitization
- **SQL Injection Protection**: Parameterized queries and ORM safety
- **Rate Limiting**: Request throttling and abuse prevention
- **Performance Optimization**: Efficient queries and connection pooling

#### **Development Tools**
- **Hot Reload**: Automatic server restart during development
- **Environment Configuration**: Flexible environment-based settings
- **Logging System**: Comprehensive error and activity logging
- **Migration Scripts**: Database schema versioning

---

## 📱 User Interface Features

#### **Responsive Design**
- **Mobile-First**: Optimized for mobile devices
- **Cross-Platform**: Works on desktop, tablet, and mobile
- **Modern UI**: Clean, intuitive interface with Tailwind CSS
- **Accessibility**: WCAG compliant design patterns

#### **Interactive Elements**
- **Real-time Chat**: Live messaging with typing indicators
- **Modal Dialogs**: Rich forms for data entry and editing
- **Drag & Drop**: File uploads and media management
- **Auto-complete**: Smart suggestions for products and services

#### **Visual Feedback**
- **Status Indicators**: Real-time order and booking status
- **Progress Tracking**: Visual progress bars and completion states
- **Notification System**: Toast notifications and alerts
- **Loading States**: Skeleton screens and spinners

---

## 🔄 Message Types & Data Flow

### **Message Object System**
- **Unified Messaging**: Text, orders, bookings as unified message objects
- **Chronological Display**: Timeline-based conversation view
- **Rich Content**: Support for media, documents, and structured data
- **Edit Capabilities**: In-line editing of orders and bookings
- **Status Updates**: Real-time status synchronization

### **Order Processing**
- **Dynamic Quantities**: Real-time quantity updates and price recalculation
- **Product Association**: Automatic product lookup and details display
- **Status Workflow**: Automated order status progression
- **Payment Integration**: Multi-payment method support
- **Delivery Tracking**: Address and delivery date management

### **Booking System**
- **Service Scheduling**: Time-slot based booking system
- **Availability Management**: Real-time service provider availability
- **Conflict Detection**: Automatic double-booking prevention
- **Reminder System**: Automated booking confirmations
- **Rescheduling**: Easy date/time modification

---

## 🔐 Authentication & Authorization

### **Multi-Role System**
- **Business Users**: Full business management access
- **Client Users**: Secure client-specific features
- **API Security**: JWT-based API authentication
- **Session Management**: Automatic token refresh and validation

### **Permission Management**
- **Business Permissions**: Full CRUD operations on business data
- **Client Permissions**: Limited to own data and conversations
- **Route Protection**: Middleware-based access control
- **Data Isolation**: Strict data separation by user role

---

## 📊 Data Models & Relationships

### **Core Entities**
- **Business**: Business profile and settings
- **Client**: Customer information and authentication
- **Conversation**: Chat sessions and message history
- **Order**: Product orders with items and payments
- **Booking**: Service appointments with scheduling
- **Product**: Catalog items with inventory
- **Service**: Bookable services with pricing

### **Relationships**
- **Business ↔ Clients**: One-to-many client relationships
- **Business ↔ Orders**: Order tracking per business
- **Client ↔ Conversations**: Communication history
- **Order ↔ OrderItems**: Detailed order line items
- **Booking ↔ Services**: Service appointment details

---

## 🚀 Deployment & Configuration

### **Environment Support**
- **Development**: SQLite database with local file storage
- **Production**: PostgreSQL with connection pooling
- **Docker Ready**: Containerized deployment support
- **Cloud Compatible**: AWS, Google Cloud, Azure deployment

### **Configuration Management**
- **Environment Variables**: Secure configuration management
- **Database Migration**: Automatic schema updates
- **Asset Management**: Static file serving and optimization
- **SSL/TLS Support**: HTTPS and secure communication

---

## Bugs
- ~~Orders and booking are loaded twice in business chat~~ ✅ **FIXED** - Removed duplicate order/booking loading from backend GetMessages function, now only MessageObjs populate frontend

## 📈 Upcoming Features

## Planning
 - order/service => payment tracking
   flow: buzz creates 
         order/service modal appears in client chat 
         client confirms the order/service
         buzz initiate payment via quick action buttons
         the payment invoice modal card will contain order/service id, category order/service, name and price
         after
         client confirms payment the buzz approve payments
        
   
### **Enhanced Communication**
- **File Sharing**: Multi-format file sharing (images, documents, media)
- **Business Products Page**: Detailed product catalog with images
- **Business Services Page**: Comprehensive service listings with descriptions
- **Advanced Search**: Full-text search across products, services, and conversations

### **Business Intelligence**
- **Analytics Dashboard**: Advanced business metrics and insights
- **Customer Analytics**: Detailed customer behavior analysis
- **Revenue Reports**: Comprehensive financial reporting
- **Performance Metrics**: KPI tracking and business health indicators

---

## 🛠️ Technical Stack

### **Backend Technologies**
- **Go**: High-performance backend language
- **Gin Framework**: Lightweight HTTP web framework
- **GORM**: Powerful ORM with database abstraction
- **JWT**: Secure authentication and authorization
- **PostgreSQL**: Production-grade relational database
- **SQLite**: Development-friendly file database

### **Frontend Technologies**
- **HTML5**: Modern semantic markup
- **Tailwind CSS**: Utility-first CSS framework
- **JavaScript**: Interactive client-side functionality
- **HTMX**: Dynamic content updates without page reloads
- **Font Awesome**: Professional icon library

### **Development Tools**
- **Hot Reload**: Live development server
- **Environment Management**: Flexible configuration system
- **Migration Tools**: Database schema versioning
- **Testing Framework**: Comprehensive test coverage

---

## 📋 Installation & Setup

### **Prerequisites**
- Go 1.19+ installed
- PostgreSQL 12+ (for production)
- Git for version control

### **Quick Start**
```bash
# Clone the repository
git clone https://github.com/your-username/threadly.git
cd threadly

# Install dependencies
go mod download

# Set up environment (copy .env.example to .env)
cp .env.example .env

# Run development server
go run cmd/server/main.go
```

### **Database Setup**
```bash
# PostgreSQL Setup (Production)
sudo -u postgres createdb threadly
sudo -u postgres createuser threadly_user

# Configure environment variables
export DB_HOST=localhost
export DB_USER=threadly_user
export DB_PASSWORD=your_password
export DB_NAME=threadly
export DB_PORT=5432
```

---

## 🤝 Contributing & Support

Threadly is an open-source project designed to help businesses manage customer relationships through modern chat interfaces. We welcome contributions and feedback!

### **Code Quality**
- Follow Go best practices and conventions
- Maintain test coverage above 80%
- Use meaningful commit messages
- Document new features and changes

### **Feature Requests**
- Open GitHub issues for new feature ideas
- Submit pull requests for improvements
- Report bugs with detailed reproduction steps
- **Suggest enhancements for user experience

---

*Threadly: Transforming customer relationships through intelligent communication and management tools.*
