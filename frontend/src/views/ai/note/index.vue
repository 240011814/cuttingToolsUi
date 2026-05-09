<script setup lang="ts">
import { h, onMounted, ref, computed } from 'vue';
import { NButton, NPopconfirm, useMessage, NTag } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchDeleteNote, fetchGetNoteList, fetchUpdateNote } from '@/service/api';
import MarkdownIt from 'markdown-it';
import { $t } from '@/locales';

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
      title: $t('page.ai.note.noteTitle'),
      key: 'title',
      width: 200,
      render(row) {
        return h('span', { class: 'font-bold' }, row.title || $t('page.ai.note.noTitle'));
      }
    },
    {
      title: $t('page.ai.note.category'),
      key: 'category',
      width: 120,
      render(row) {
        return h(NTag, { type: 'info', bordered: false }, { default: () => row.category });
      }
    },
    {
      title: $t('page.ai.note.content'),
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
      title: $t('page.ai.note.createdAt'),
      key: 'createdAt',
      width: 180,
      render(row) {
        return h('span', new Date(row.createdAt).toLocaleString());
      }
    },
    {
      title: $t('page.ai.note.actions'),
      key: 'actions',
      width: 200,
      fixed: 'right',
      render(row) {
        return h('div', { class: 'flex gap-2' }, [
          h(NButton, { size: 'small', type: 'primary', quaternary: true, onClick: () => handleView(row) }, { default: () => $t('page.ai.note.view') }),
          h(NButton, { size: 'small', type: 'info', quaternary: true, onClick: () => handleEdit(row) }, { default: () => $t('page.ai.note.edit') }),
          h(
          NPopconfirm,
          {
            onPositiveClick: () => handleDelete(row.id),
            trigger: 'click'
          },
          {
            trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => $t('common.delete') }),
            default: () => $t('page.ai.note.deleteConfirm')
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
    message.error(`${$t('page.ai.note.loadFailed')}: ${err?.message || $t('common.error')}`);
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (id: number) => {
  try {
    await fetchDeleteNote(id);
    message.success($t('page.ai.note.deleteSuccess'));
    if (data.value.length === 1 && pagination.value.page > 1) {
      pagination.value.page--;
    }
    loadData();
  } catch (err: any) {
    message.error(`${$t('page.ai.note.deleteFailed')}: ${err?.message || $t('common.error')}`);
  }
};

const showViewModal = ref(false);
const currentNoteContent = ref('');

const showEditModal = ref(false);
const isEdit = ref(false);
const editLoading = ref(false);
const editForm = ref({ id: 0, title: '', category: '', content: '' });

const handleAdd = () => {
  isEdit.value = false;
  editForm.value = { id: 0, title: '', category: '', content: '' };
  showEditModal.value = true;
};

const handleView = (row: any) => {
  currentNoteContent.value = row.content || '';
  showViewModal.value = true;
};

const handleEdit = (row: any) => {
  isEdit.value = true;
  editForm.value = { id: row.id, title: row.title, category: row.category, content: row.content };
  showEditModal.value = true;
};

const submitEdit = async () => {
  if (!editForm.value.title.trim() || !editForm.value.category.trim() || !editForm.value.content.trim()) {
    message.warning($t('page.ai.note.fieldsRequired'));
    return;
  }
  editLoading.value = true;
  try {
    if (isEdit.value) {
      await fetchUpdateNote(editForm.value.id, {
        title: editForm.value.title,
        category: editForm.value.category,
        content: editForm.value.content
      });
      message.success($t('page.ai.note.updateSuccess'));
    } else {
      const { fetchAddNote } = await import('@/service/api');
      await fetchAddNote({
        title: editForm.value.title,
        category: editForm.value.category,
        content: editForm.value.content
      });
      message.success($t('page.ai.note.addSuccess'));
    }
    showEditModal.value = false;
    loadData();
  } catch (err: any) {
    message.error(`${$t('page.ai.note.operationFailed')}: ${err?.message || $t('common.error')}`);
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
          <span class="text-18px font-bold">{{ $t('page.ai.note.title') }}</span>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <NInput
              v-model:value="keyword"
              :placeholder="$t('page.ai.note.searchTitlePlaceholder')"
              clearable
              style="width: 260px"
              @keyup.enter="loadData"
            />
            <NInput
              v-model:value="category"
              :placeholder="$t('page.ai.note.searchCategoryPlaceholder')"
              clearable
              style="width: 200px"
              @keyup.enter="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <icon-mdi-magnify class="text-icon" />
              </template>
              {{ $t('common.search') }}
            </NButton>
          </div>
          <div class="flex gap-2 items-center">
            <ButtonIcon
              icon="mdi:refresh"
              :tooltip-content="$t('common.refresh')"
              @click="loadData"
            />
            <NButton type="primary" @click="handleAdd">
              <template #icon>
                <icon-mdi-plus class="text-icon" />
              </template>
              {{ $t('page.ai.note.addNote') }}
            </NButton>
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
      :title="$t('page.ai.note.viewNote')"
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
      :title="isEdit ? $t('page.ai.note.editNote') : $t('page.ai.note.addNote')"
      style="width: 800px; max-width: 95vw;"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="editForm" label-placement="left" label-width="80">
        <div class="flex gap-4">
          <NFormItem :label="$t('page.ai.note.noteTitle')" path="title" class="flex-1">
            <NInput v-model:value="editForm.title" :placeholder="$t('page.ai.note.noteTitlePlaceholder')" />
          </NFormItem>
          <NFormItem :label="$t('page.ai.note.category')" path="category" style="width: 240px">
            <NInput v-model:value="editForm.category" :placeholder="$t('page.ai.note.searchCategoryPlaceholder')" />
          </NFormItem>
        </div>
        <NFormItem :label="$t('page.ai.note.content')" path="content">
          <div class="grid grid-cols-2 gap-4 w-full">
            <NInput
              v-model:value="editForm.content"
              type="textarea"
              :autosize="{ minRows: 12, maxRows: 15 }"
              :placeholder="$t('page.ai.note.noteContentPlaceholder')"
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
          <NButton @click="showEditModal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="editLoading" @click="submitEdit">
            {{ $t('page.ai.note.saveSuccess') }}
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped></style>
