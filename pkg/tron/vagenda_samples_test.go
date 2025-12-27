package tron

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestVAgendaDocSamples(t *testing.T) {
	tests := []struct {
		name string
		tron string
		json string
	}{
		{
			name: "minimal-todolist",
			tron: `class vAgendaInfo: version
class TodoList: items
class TodoItem: title, status

vAgendaInfo: vAgendaInfo("0.2")
todoList: TodoList([
  TodoItem("Implement authentication", "pending"),
  TodoItem("Write API documentation", "pending")
])
`,
			json: `{
  "vAgendaInfo": {
    "version": "0.2"
  },
  "todoList": {
    "items": [
      {
        "title": "Implement authentication",
        "status": "pending"
      },
      {
        "title": "Write API documentation",
        "status": "pending"
      }
    ]
  }
}`,
		},
		{
			name: "minimal-plan",
			tron: `class vAgendaInfo: version
class Plan: title, status, narratives, phases
class Phase: title, status
class Narrative: title, content

vAgendaInfo: vAgendaInfo("0.2")
plan: Plan(
  "Add user authentication",
  "draft",
  {
    "proposal": Narrative(
      "Proposed Changes",
      "Implement JWT-based authentication with refresh tokens"
    )
  },
  [
    Phase("Database schema", "completed"),
    Phase("JWT implementation", "pending")
  ]
)
`,
			json: `{
  "vAgendaInfo": {
    "version": "0.2"
  },
  "plan": {
    "title": "Add user authentication",
    "status": "draft",
    "narratives": {
      "proposal": {
        "title": "Proposed Changes",
        "content": "Implement JWT-based authentication with refresh tokens"
      }
    },
    "phases": [
      {
        "title": "Database schema",
        "status": "completed"
      },
      {
        "title": "JWT implementation",
        "status": "pending"
      }
    ]
  }
}`,
		},
		{
			name: "version-control-and-sync",
			tron: `class vAgendaInfo: version
class TodoList: id, items, uid, agent, sequence, changeLog
class TodoItem: id, title, status
class Agent: id, type, name, model
class Change: sequence, timestamp, agent, operation, reason

vAgendaInfo: vAgendaInfo("0.2")
todoList: TodoList(
  "todo-002",
  [
    TodoItem("item-8", "Sync tasks across devices", "completed")
  ],
  "550e8400-e29b-41d4-a716-446655440000",
  Agent("agent-1", "aiAgent", "Claude", "claude-3.5-sonnet"),
  3,
  [
    Change(1, "2024-12-27T10:00:00Z", Agent("agent-1", "aiAgent", "Claude", null), "create", "Initial creation"),
    Change(2, "2024-12-27T10:30:00Z", Agent("agent-1", "aiAgent", "Claude", null), "update", "Added new item"),
    Change(3, "2024-12-27T11:00:00Z", Agent("agent-1", "aiAgent", "Claude", null), "update", "Marked item completed")
  ]
)
`,
			json: `{
  "vAgendaInfo": {
    "version": "0.2"
  },
  "todoList": {
    "id": "todo-002",
    "items": [
      {
        "id": "item-8",
        "title": "Sync tasks across devices",
        "status": "completed"
      }
    ],
    "uid": "550e8400-e29b-41d4-a716-446655440000",
    "agent": {
      "id": "agent-1",
      "type": "aiAgent",
      "name": "Claude",
      "model": "claude-3.5-sonnet"
    },
    "sequence": 3,
    "changeLog": [
      {
        "sequence": 1,
        "timestamp": "2024-12-27T10:00:00Z",
        "agent": {
          "id": "agent-1",
          "type": "aiAgent",
          "name": "Claude",
          "model": null
        },
        "operation": "create",
        "reason": "Initial creation"
      },
      {
        "sequence": 2,
        "timestamp": "2024-12-27T10:30:00Z",
        "agent": {
          "id": "agent-1",
          "type": "aiAgent",
          "name": "Claude",
          "model": null
        },
        "operation": "update",
        "reason": "Added new item"
      },
      {
        "sequence": 3,
        "timestamp": "2024-12-27T11:00:00Z",
        "agent": {
          "id": "agent-1",
          "type": "aiAgent",
          "name": "Claude",
          "model": null
        },
        "operation": "update",
        "reason": "Marked item completed"
      }
    ]
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var want interface{}
			if err := json.Unmarshal([]byte(tt.json), &want); err != nil {
				t.Fatalf("invalid JSON fixture: %v", err)
			}

			var got interface{}
			if err := Unmarshal([]byte(tt.tron), &got); err != nil {
				t.Fatalf("failed to unmarshal TRON sample: %v", err)
			}

			if !reflect.DeepEqual(normalizeJSONValue(want), normalizeJSONValue(got)) {
				t.Fatalf("decoded mismatch\nwant: %#v\ngot: %#v", want, got)
			}
		})
	}
}
