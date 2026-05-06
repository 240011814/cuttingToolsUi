<script setup lang="ts">
import { h, onMounted, ref, computed } from 'vue';
import { NButton, NPopconfirm, useMessage, NTag } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchDeleteNote, fetchGetNoteList, fetchUpdateNote } from '@/service/api';
import MarkdownIt from 'markdown-it';

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
});

const message = useMessage();
const loading = ref(false);
const data = ref<any[]>([]);
const keyword = ref('');
const category = ref('');
const total = ref(0);
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.value.page = page;
    loadData();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    pagination.value.page = 1;
    loadData();
  }
});

const columns = computed<DataTableColumns<any>>(() => {
  return [
    {
      title: '分类',
      key: 'category',
      width: 150,
      render(row) {
        return h(NTag, { type: 'info', bordered: false }, { default: () => row.category });
      }
    },
    {
      title: '内容',
      key: 'content',
      minWidth: 300,
      render(row) {
        return h('div', {
          class: 'prose dark:prose-invert max-w-none max-h-24 overflow-y-auto p-2 bg-gray-50/50 dark:bg-dark-100 rounded-md text-sm leading-relaxed',
          innerHTML: md.render(row.content || '')
        });
      }
    },
    {
      title: '创建时间',
      key: 'createdAt',
      width: 180,
      render(row) {
        return h('span', new Date(row.createdAt).toLocaleString());
      }
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render(row) {
        return h('div', { class: 'flex gap-2' }, [
          h(NButton, { size: 'small', type: 'primary', quaternary: true, onClick: () => handleView(row) }, { default: () => '查看' }),
          h(NButton, { size: 'small', type: 'info', quaternary: true, onClick: () => handleEdit(row) }, { default: () => '编辑' }),
          h(
          NPopconfirm,
          {
            onPositiveClick: () => handleDelete(row.id),
            trigger: 'click'
          },
          {
            trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => '删除' }),
            default: () => '确定删除此笔记吗？'
          }
        )
        ]);
      }
    }
  ];
});

const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchGetNoteList({
      content: keyword.value,
      category: category.value,
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });
    if (res) {
      data.value = res.items;
      total.value = res.total;
      pagination.value.itemCount = res.total;
    }
  } catch (err: any) {
    message.error(`获取列表失败: ${err?.message || '未知错误'}`);
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (id: number) => {
  try {
    await fetchDeleteNote(id);
    message.success('删除成功');
    if (data.value.length === 1 && pagination.value.page > 1) {
      pagination.value.page--;
    }
    loadData();
  } catch (err: any) {
    message.error(`删除失败: ${err?.message || '未知错误'}`);
  }
};

const showViewModal = ref(false);
const currentNoteContent = ref('');

const showEditModal = ref(false);
const editLoading = ref(false);
const editForm = ref({ id: 0, category: '', content: '' });

const handleView = (row: any) => {
  currentNoteContent.value = row.content || '';
  showViewModal.value = true;
};

const handleEdit = (row: any) => {
  editForm.value = { id: row.id, category: row.category, content: row.content };
  showEditModal.value = true;
};

const submitEdit = async () => {
  if (!editForm.value.category.trim() || !editForm.value.content.trim()) {
    message.warning('分类和内容不能为空');
    return;
  }
  editLoading.value = true;
  try {
    await fetchUpdateNote(editForm.value.id, {
      category: editForm.value.category,
      content: editForm.value.content
    });
    message.success('修改成功');
    showEditModal.value = false;
    loadData();
  } catch (err: any) {
    message.error(`修改失败: ${err?.message || '未知错误'}`);
  } finally {
    editLoading.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center gap-4">
          <span class="text-18px font-bold">笔记本</span>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <NInput
              v-model:value="keyword"
              placeholder="搜索内容..."
              clearable
              style="width: 260px"
              @keyup.enter="loadData"
            />
            <NInput
              v-model:value="category"
              placeholder="搜索分类..."
              clearable
              style="width: 200px"
              @keyup.enter="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <icon-mdi-magnify class="text-icon" />
              </template>
              搜索
            </NButton>
          </div>
          <div class="flex gap-2 items-center">
            <ButtonIcon
              icon="mdi:refresh"
              tooltip-content="刷新"
              @click="loadData"
            />
          </div>
        </div>

        <NDataTable
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="pagination"
          remote
          :row-key="(row) => row.id"
          flex-height
          class="flex-1"
        />
      </div>
    </NCard>

    <!-- 查看模态框 -->
    <NModal
      v-model:show="showViewModal"
      preset="card"
      title="查看笔记"
      style="width: 800px; max-width: 95vw;"
      :segmented="{ content: 'soft' }"
    >
      <div
        class="prose dark:prose-invert max-w-none overflow-y-auto p-4 bg-gray-50/50 dark:bg-dark-100 rounded-md text-sm leading-relaxed"
        style="max-height: 70vh;"
        v-html="md.render(currentNoteContent)"
      ></div>
    </NModal>

    <!-- 编辑模态框 -->
    <NModal
      v-model:show="showEditModal"
      preset="card"
      title="编辑笔记"
      style="width: 800px; max-width: 95vw;"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="editForm" label-placement="left" label-width="80">
        <NFormItem label="分类" path="category">
          <NInput v-model:value="editForm.category" placeholder="输入笔记分类" style="width: 200px" />
        </NFormItem>
        <NFormItem label="内容" path="content">
          <div class="grid grid-cols-2 gap-4 w-full">
            <NInput
              v-model:value="editForm.content"
              type="textarea"
              :autosize="{ minRows: 12, maxRows: 15 }"
              placeholder="输入笔记内容"
            />
            <div
              class="prose dark:prose-invert max-w-none overflow-y-auto p-4 border border-gray-200 dark:border-gray-700 rounded-md bg-gray-50/50 dark:bg-dark-100 text-sm leading-relaxed"
              style="height: 100%; max-height: 350px;"
              v-html="md.render(editForm.content)"
            ></div>
          </div>
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showEditModal = false">取消</NButton>
          <NButton type="primary" :loading="editLoading" @click="submitEdit">
            保存修改
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped></style>
