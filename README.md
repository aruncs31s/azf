etlabauthzframework/README.md
# AZF Authorization Framework Documentation

Welcome to the comprehensive documentation for the AZF Authorization Framework! This framework provides enterprise-grade role-based access control (RBAC) and authorization management for your applications.

## 📚 Documentation Structure

### Getting Started
- **[Quick Start](QUICKSTART.md)** - Get up and running in 5 minutes
- **[Installation](guides/installation.md)** - Detailed installation instructions
- **[Basic Usage](guides/basic-usage.md)** - Learn the fundamentals

### Guides
- **[Authentication](guides/authentication.md)** - User authentication setup
- **[Authorization](guides/authorization.md)** - Role and permission management
- **[API Usage](guides/api-usage.md)** - RESTful API endpoints

### Examples
- **[Basic Setup](examples/basic-setup.md)** - Simple implementation example
- **[Advanced Usage](examples/advanced-usage.md)** - Complex scenarios
- **[Custom Policies](examples/custom-policies.md)** - Policy customization

## 🎯 Key Features

- **Role-Based Access Control (RBAC)** ✅ - Manage user roles and permissions using Casbin
- **Enterprise Authorization** ✅ - Advanced policy management with route metadata
- **RESTful API** ✅ - Easy-to-use HTTP endpoints for management
- **Dashboard UI** ✅ - Visual management interface with modern design
- **Audit Logging** ✅ - Track all authorization events and API usage
- **JWT Support** ✅ - Secure token-based authentication
- **API Analytics** ✅ - Real-time monitoring of API performance and usage
- **Route Metadata Management** ✅ - Comprehensive endpoint configuration
- **Dark Mode UI** ✅ - Automatic theme switching with responsive design
- **Component Library** ✅ - Reusable UI components with template inheritance
- **Ownership Checks** ✅ - Resource-level access control
- **Usage Tracking** ✅ - Automatic API request monitoring

## 🚀 Quick Links

| Feature | Documentation | Status |
|---------|---|--------|
| Create Roles | [Role Management Guide](guides/authentication.md) | ✅ Available |
| Assign Permissions | [Authorization Guide](guides/authorization.md) | ✅ Available |
| API Reference | [FEATURES.md](FEATURES.md) | ✅ Complete |
| Admin Dashboard | Access at `/admin-ui` after login | ✅ Production Ready |
| Route Metadata | [Route Metadata Quick Start](docs/ROUTE_METADATA_QUICK_START.md) | ✅ Complete |
| UI Framework | [UI Architecture Guide](docs/UI_ARCHITECTURE.md) | ✅ Complete |

## 💡 Common Tasks

### Create a New Role
```go
// POST /admin-ui/api/roles
{
  "name": "editor",
  "description": "Can edit content"
}
```

### Assign Role to User
```go
// POST /admin-ui/api/roles/assign
{
  "user_id": "user123",
  "role": "editor"
}
```

### Check Permissions
```go
// Use middleware to protect routes
router.GET("/protected", authMiddleware, handler)
```

### Configure Route Metadata
```json
{
  "path": "/api/users",
  "method": "GET",
  "description": "Get users",
  "allowed_roles": ["admin"],
  "ownership_check": true,
  "audit_required": true
}
```

## 📖 API Overview

The framework provides RESTful endpoints for:
- User authentication and JWT management
- Role management and assignment
- Permission assignment and validation
- Route metadata configuration
- Audit log retrieval and analytics
- API usage tracking and performance monitoring

For detailed API documentation, see **[FEATURES.md](FEATURES.md)**

## 🔒 Security

This framework implements industry best practices:
- JWT token validation with configurable secrets
- Role-based access control via Casbin
- Route metadata validation and ownership checks
- Audit trail logging for compliance
- CORS protection and input validation
- Admin authentication with session management

## 🛠️ Admin Dashboard

Access the admin dashboard at `http://localhost:8080/admin-ui/login`

**Features:**
- **Home Dashboard**: System overview with real-time statistics
- **API Analytics**: Performance monitoring and usage trends
- **Route Metadata Management**: Full CRUD operations for endpoint configuration
- **Role Management**: View and manage roles and user assignments
- **Audit Logs**: Comprehensive authorization event tracking
- **Policy Management**: Casbin policy viewing and management

**UI Features:**
- Modern responsive design with Tailwind CSS
- Automatic dark mode support
- Component-based architecture with template inheritance
- Real-time updates and interactive charts

## 📝 API Endpoints

### Authentication
- `POST /admin-ui/login/json` - JWT token generation
- `GET /admin-ui/logout` - Session cleanup

### Route Management
- `GET /admin-ui/route_metadata` - View all routes
- `POST /admin-ui/route_metadata` - Save route metadata
- `POST /admin-ui/route_metadata/import` - Import routes from JSON
- `POST /admin-ui/route_metadata/delete` - Delete route

### Analytics
- `GET /admin-ui/api_analytics` - API usage dashboard
- `GET /admin-ui/api_analytics/endpoint` - Endpoint details

### Roles & Policies
- `GET /admin-ui/roles` - Role management interface
- `POST /admin-ui/api/roles` - Create roles
- `PUT /admin-ui/api/roles` - Update roles
- `POST /admin-ui/api/roles/assign` - Assign roles to users

## 📊 Monitoring

Monitor your authorization system with:
- **Real-time API Analytics Dashboard** - Track endpoint performance
- **Audit Log Viewer** - Review authorization decisions
- **Usage Statistics** - Monitor API calls and trends
- **Performance Metrics** - Response times and error rates

Access these at `/admin-ui/api_analytics` and `/admin-ui/audit_logs`

## ❓ FAQ

**Q: How do I create a new role?**  
A: Use the admin dashboard at `/admin-ui/roles` or POST to `/admin-ui/api/roles`

**Q: Can I use custom permissions?**  
A: Yes, permissions are fully customizable via Casbin policies and route metadata

**Q: Is JWT required?**  
A: JWT is the recommended authentication method, but the framework is flexible

**Q: How are permissions enforced?**  
A: Through middleware on protected routes, with Casbin policy evaluation

**Q: Can I customize the UI?**  
A: Yes, the UI framework supports template inheritance and component extension

## 🤝 Contributing

To contribute to the documentation:
1. Edit markdown files in the `/docs` folder
2. Follow the folder structure
3. Use clear, concise language
4. Add examples where helpful

## 📞 Support

For questions or issues:
1. Check the troubleshooting guides
2. Review example implementations
3. Consult the API documentation
4. Check audit logs for error details

## 🎓 Learning Path

**Beginner:**
1. Read [QUICKSTART.md](QUICKSTART.md)
2. Follow [Basic Usage Guide](guides/basic-usage.md)
3. Try [Basic Setup Example](examples/basic-setup.md)

**Intermediate:**
1. Study [Authentication Guide](guides/authentication.md)
2. Learn [Authorization](guides/authorization.md)
3. Explore [API Usage](guides/api-usage.md)

**Advanced:**
1. Implement [Advanced Usage](examples/advanced-usage.md)
2. Create [Custom Policies](examples/custom-policies.md)
3. Customize the [UI Framework](docs/UI_ARCHITECTURE.md)

## 📋 Latest Documentation

- ✅ Core framework documentation complete
- ✅ API reference available
- ✅ Example implementations included
- ✅ Troubleshooting guides provided
- ✅ UI framework architecture documented
- ✅ Route metadata management complete
- ✅ Component library documented
- ✅ Implementation summaries available

## 🔄 Updates

The documentation is regularly updated to reflect framework changes. Check back often for new guides and examples.

---

**Framework Version:** 1.0.0  
**Last Updated:** 2025-11-14  
**Status:** Production Ready  
**Go Version:** 1.25.3+  

---

*Built with Go, Casbin, Gin, and modern web technologies*