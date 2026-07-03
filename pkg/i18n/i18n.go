package i18n

import "github.com/gofiber/fiber/v2"

const (
	EN = "en"
	TH = "th"
)

var messages = map[string]map[string]string{
	// Auth
	"auth.login_success":    {EN: "logged in successfully", TH: "เข้าสู่ระบบสำเร็จ"},
	"auth.register_success": {EN: "user registered", TH: "ลงทะเบียนสำเร็จ"},

	// POS Products
	"product.list":   {EN: "products retrieved", TH: "ดึงข้อมูลสินค้าสำเร็จ"},
	"product.get":    {EN: "product retrieved", TH: "ดึงข้อมูลสินค้าสำเร็จ"},
	"product.create": {EN: "product created", TH: "สร้างสินค้าสำเร็จ"},
	"product.update": {EN: "product updated", TH: "อัปเดตสินค้าสำเร็จ"},
	"product.delete": {EN: "product deleted", TH: "ลบสินค้าสำเร็จ"},

	// Orders
	"order.created":   {EN: "order created", TH: "สร้างออเดอร์สำเร็จ"},
	"order.list":      {EN: "orders retrieved", TH: "ดึงข้อมูลออเดอร์สำเร็จ"},
	"order.get":       {EN: "order retrieved", TH: "ดึงข้อมูลออเดอร์สำเร็จ"},
	"order.cancelled": {EN: "order cancelled", TH: "ยกเลิกออเดอร์แล้ว"},
	"order.paid":      {EN: "marked as paid", TH: "บันทึกการชำระเงินแล้ว"},

	// Stock
	"stock.list":         {EN: "stock retrieved", TH: "ดึงข้อมูลสต็อกสำเร็จ"},
	"stock.synced":       {EN: "sync complete", TH: "ซิงค์สำเร็จ"},
	"stock.availability": {EN: "availability checked", TH: "ตรวจสอบสต็อกสำเร็จ"},
	"stock.barcode":      {EN: "product retrieved", TH: "ดึงข้อมูลสินค้าสำเร็จ"},

	// Config
	"config.bank_qr_get":     {EN: "bank QR config retrieved", TH: "ดึงข้อมูลการตั้งค่า QR สำเร็จ"},
	"config.bank_qr_updated": {EN: "bank QR config updated", TH: "อัปเดตการตั้งค่า QR สำเร็จ"},
	"config.vat_get":         {EN: "vat config retrieved", TH: "ดึงข้อมูลการตั้งค่า VAT สำเร็จ"},
	"config.vat_updated":     {EN: "vat config updated", TH: "อัปเดตการตั้งค่า VAT สำเร็จ"},

	// Reports
	"report.summary":          {EN: "report summary", TH: "สรุปรายงาน"},
	"report.daily_revenue":    {EN: "daily revenue", TH: "รายได้รายวัน"},
	"report.top_products":     {EN: "top products", TH: "สินค้าขายดี"},
	"report.category_revenue": {EN: "revenue by category", TH: "รายได้ตามหมวดหมู่"},
	"report.cashier_sales":    {EN: "cashier sales", TH: "ยอดขายต่อแคชเชียร์"},
	"report.overdue_paylater": {EN: "overdue pay-later orders", TH: "ออเดอร์ค้างชำระ"},

	// Validation / error keys
	"err.invalid_body":                 {EN: "invalid request body", TH: "รูปแบบคำขอไม่ถูกต้อง"},
	"err.email_password_required":      {EN: "email and password are required", TH: "กรุณากรอกอีเมลและรหัสผ่าน"},
	"err.invalid_credentials":          {EN: "invalid credentials", TH: "อีเมลหรือรหัสผ่านไม่ถูกต้อง"},
	"err.account_inactive":             {EN: "account is inactive", TH: "บัญชีนี้ถูกระงับการใช้งาน"},
	"err.email_exists":                 {EN: "email already registered", TH: "อีเมลนี้ถูกใช้งานแล้ว"},
	"err.register_fields_required":     {EN: "name, email, and password are required", TH: "กรุณากรอกชื่อ อีเมล และรหัสผ่าน"},
	"err.password_too_short":           {EN: "password must be at least 6 characters", TH: "รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร"},
	"err.product_not_found":            {EN: "product not found", TH: "ไม่พบสินค้า"},
	"err.pos_product_id_name_required": {EN: "pos_product_id and name are required", TH: "กรุณากรอก pos_product_id และชื่อสินค้า"},
	"err.price_non_negative":           {EN: "price must be non-negative", TH: "ราคาต้องไม่ติดลบ"},
	"err.pos_product_id_exists":        {EN: "pos_product_id already exists", TH: "pos_product_id นี้มีอยู่แล้ว"},
	"err.order_not_found":              {EN: "order not found", TH: "ไม่พบออเดอร์"},
	"err.bank_qr_not_configured":       {EN: "bank QR config not configured", TH: "ยังไม่ได้ตั้งค่า Bank QR"},
	"err.bank_qr_fields_required":      {EN: "bank_name, account_name, and account_number are required", TH: "กรุณากรอก bank_name, account_name และ account_number"},
	"err.promptpay_not_set":            {EN: "promptpay_id not set in bank QR config — update it via PUT /api/v1/config/bank-qr", TH: "ยังไม่ได้ตั้งค่า promptpay_id — กรุณาอัปเดตผ่าน PUT /api/v1/config/bank-qr"},
	"err.qr_generation_failed":         {EN: "failed to generate QR code", TH: "สร้าง QR Code ไม่สำเร็จ"},
	"err.inventory_sync_failed":        {EN: "inventory sync failed", TH: "ซิงค์สต็อกไม่สำเร็จ"},
	"err.inventory_check_failed":       {EN: "inventory check failed", TH: "ตรวจสอบสต็อกไม่สำเร็จ"},
	"err.inventory_lookup_failed":      {EN: "inventory lookup failed", TH: "ค้นหาสินค้าจากระบบสต็อกไม่สำเร็จ"},
	"err.stock_deduction_failed":       {EN: "stock deduction failed", TH: "หักสต็อกไม่สำเร็จ กรุณาตรวจสอบสินค้าในระบบคลัง"},

	// Service error strings (bubbled up from service layer)
	"invalid credentials":                    {EN: "invalid credentials", TH: "อีเมลหรือรหัสผ่านไม่ถูกต้อง"},
	"account is inactive":                    {EN: "account is inactive", TH: "บัญชีนี้ถูกระงับการใช้งาน"},
	"email already registered":               {EN: "email already registered", TH: "อีเมลนี้ถูกใช้งานแล้ว"},
	"name, email, and password are required": {EN: "name, email, and password are required", TH: "กรุณากรอกชื่อ อีเมล และรหัสผ่าน"},
	"password must be at least 6 characters": {EN: "password must be at least 6 characters", TH: "รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร"},
	"product not found":                      {EN: "product not found", TH: "ไม่พบสินค้า"},
	"pos_product_id and name are required":   {EN: "pos_product_id and name are required", TH: "กรุณากรอก pos_product_id และชื่อสินค้า"},
	"price must be non-negative":             {EN: "price must be non-negative", TH: "ราคาต้องไม่ติดลบ"},
	"pos_product_id already exists":          {EN: "pos_product_id already exists", TH: "pos_product_id นี้มีอยู่แล้ว"},
	"cost_price must be non-negative":        {EN: "cost_price must be non-negative", TH: "ต้นทุนต้องไม่ติดลบ"},

	// Order service error strings
	"order not found":                   {EN: "order not found", TH: "ไม่พบออเดอร์"},
	"order must have at least one item": {EN: "order must have at least one item", TH: "ออเดอร์ต้องมีสินค้าอย่างน้อย 1 รายการ"},
	"customer_name and customer_phone are required for PAY_LATER": {EN: "customer_name and customer_phone are required for PAY_LATER", TH: "กรุณากรอก customer_name และ customer_phone สำหรับการชำระภายหลัง"},
	"only PENDING orders can be cancelled":                        {EN: "only PENDING orders can be cancelled", TH: "ยกเลิกได้เฉพาะออเดอร์ที่อยู่ในสถานะ PENDING เท่านั้น"},
	"unauthorized to cancel this order":                           {EN: "unauthorized to cancel this order", TH: "ไม่มีสิทธิ์ยกเลิกออเดอร์นี้"},
	"only PAY_LATER orders can be marked as paid":                 {EN: "only PAY_LATER orders can be marked as paid", TH: "บันทึกการชำระเงินได้เฉพาะออเดอร์ประเภท PAY_LATER เท่านั้น"},
	"order is already paid":                                       {EN: "order is already paid", TH: "ออเดอร์นี้ชำระเงินแล้ว"},
	"only COMPLETED orders can be marked as paid":                 {EN: "only COMPLETED orders can be marked as paid", TH: "บันทึกการชำระเงินได้เฉพาะออเดอร์ที่สำเร็จแล้วเท่านั้น"},

	// VAT config error strings
	"vat rate must be between 0 and 100": {EN: "vat rate must be between 0 and 100", TH: "อัตรา VAT ต้องอยู่ระหว่าง 0 ถึง 100"},
}

// T returns the translated message for the given language, falling back to EN, then the key itself.
func T(lang, key string) string {
	if m, ok := messages[key]; ok {
		if v, ok := m[lang]; ok && v != "" {
			return v
		}
		if v, ok := m[EN]; ok {
			return v
		}
	}
	return key
}

// Lang reads the language stored by the language middleware.
func Lang(c *fiber.Ctx) string {
	if lang, ok := c.Locals("lang").(string); ok && lang != "" {
		return lang
	}
	return EN
}

// Resolve detects language from ?lang= query param or Accept-Language header.
func Resolve(c *fiber.Ctx) string {
	if lang := c.Query("lang"); lang == TH || lang == EN {
		return lang
	}
	accept := c.Get("Accept-Language")
	if len(accept) >= 2 && accept[:2] == "th" {
		return TH
	}
	return EN
}
