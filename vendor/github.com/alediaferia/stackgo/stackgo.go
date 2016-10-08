// All you need to work with stackgo stacks
//
// For an example usage visit http://github.com/alediaferia/stackgo
package stackgo

type Stack struct {
	size int
	currentPage []interface{}
	pages [][]interface{}
	offset int
	capacity int
	pageSize int
	currentPageIndex int
}

const s_DefaultAllocPageSize = 4096

// NewStack Creates a new Stack object with
// an underlying default block allocation size.
// The current default allocation size is one page.
// If you want to use a different block size use
//  NewStackWithCapacity()
func NewStack() *Stack {
	stack := new(Stack)
	stack.currentPage = make([]interface{}, s_DefaultAllocPageSize)
	stack.pages = [][]interface{}{stack.currentPage}
	stack.offset = 0
	stack.capacity = s_DefaultAllocPageSize
	stack.pageSize = s_DefaultAllocPageSize
	stack.size = 0
	stack.currentPageIndex = 0

	return stack
}


// NewStackWithCapacity makes it easy to specify
// a custom block size for inner slice backing the
// stack
func NewStackWithCapacity(cap int) *Stack {
	stack := new(Stack)
	stack.currentPage = make([]interface{}, cap)
	stack.pages = [][]interface{}{stack.currentPage}
	stack.offset = 0
	stack.capacity = cap
	stack.pageSize = cap
	stack.size = 0
	stack.currentPageIndex = 0

	return stack
}


// Push pushes a new element to the stack
func (s *Stack) Push(elem... interface{}) {
	if elem == nil || len(elem) == 0 {
		return
	}

    if s.size == s.capacity {
		pages_count := len(elem) / s.pageSize
		if len(elem) % s.pageSize != 0 {
			pages_count++
		}
		s.capacity += s.pageSize

		s.currentPage = make([]interface{}, s.pageSize)
		s.pages = append(s.pages, s.currentPage)
		s.currentPageIndex++

		pages_count--
		for pages_count > 0 {
			page := make([]interface{}, s.pageSize)
			s.pages = append(s.pages, page)
		}

		s.offset = 0
	}

	available := len(s.currentPage) - s.offset
	for len(elem) > available {
		copy(s.currentPage[s.offset:], elem[:available])
		s.currentPage = s.pages[s.currentPageIndex + 1]
		s.currentPageIndex++
		elem = elem[available:]
		s.offset = 0
		available = len(s.currentPage)
	}

	copy(s.currentPage[s.offset:], elem)
	s.offset += len(elem)
	s.size += len(elem)
}

// Pop pops the top element from the stack
// If the stack is empty it returns nil
func (s *Stack) Pop() (elem interface{}) {
	if s.size == 0 {
		return nil
	}

	s.offset--
	s.size--
	if s.offset < 0 {
		s.offset = s.pageSize - 1

		s.currentPage, s.pages = s.pages[len(s.pages) - 2], s.pages[:len(s.pages) - 1]
		s.capacity -= s.pageSize
		s.currentPageIndex--
	}

	elem = s.currentPage[s.offset]

	return
}

func (s *Stack) Top() (elem interface{}) {
	if s.size == 0 {
		return nil
	}

	off := s.offset - 1
	if off < 0 {
		page := s.pages[len(s.pages)-1]
		elem = page[len(page)-1]
		return
	}
	elem = s.currentPage[off]
	return
}

// The current size of the stack
func (s *Stack) Size() int {
	return s.size
}
