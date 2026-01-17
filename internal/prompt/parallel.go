package prompt

import (
	"fmt"

	"github.com/ncecere/ralphy/internal/tasks"
)

func BuildParallel(task tasks.Task) string {
	label := taskLabel(task)
	return fmt.Sprintf(`You are working on a specific task. Focus ONLY on this task:

TASK: %s

Instructions:
1. Implement this specific task completely
2. Write tests if appropriate
3. Update progress.txt with what you did
4. Commit your changes with a descriptive message

Do NOT modify PRD.md or mark tasks complete - that will be handled separately.
Focus only on implementing: %s`, label, task.Title)
}
