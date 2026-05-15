package notifications

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "notifications",
	Description: "Push notification subscription management.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/notifications/vapid-public-key",
			Summary:      "Return VAPID public key",
			Description:  "Returns the server VAPID public key for creating a push subscription.",
			Auth:         "none",
			ResponseBody: "VAPIDPublicKeyResponse",
		},
		{
			Method:       "POST",
			Path:         "/notifications/subscriptions",
			Summary:      "Save push subscription",
			Description:  "Saves or replaces the push subscription for the authenticated user.",
			Auth:         "bearer token required",
			RequestBody:  "SaveSubscriptionRequest",
			ResponseBody: "SubscriptionResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/notifications/subscriptions",
			Summary:      "Delete push subscription",
			Description:  "Removes the push subscription for the authenticated user.",
			Auth:         "bearer token required",
			ResponseBody: "DeleteResponse",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
	},
}
